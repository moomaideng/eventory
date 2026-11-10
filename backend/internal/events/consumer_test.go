package events

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startRabbitMQ(t *testing.T) string {
	t.Helper()
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx := context.Background()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "rabbitmq:4-alpine",
			ExposedPorts: []string{"5672/tcp"},
			WaitingFor:   wait.ForLog("Server startup complete").WithStartupTimeout(90 * time.Second),
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	endpoint, err := container.PortEndpoint(ctx, "5672/tcp", "")
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("amqp://guest:guest@%s/", endpoint)
}

type recorder struct {
	mu    sync.Mutex
	calls map[string]int // by body
}

// name decodes the JSON string the tests publish as a message body.
func name(body []byte) string {
	var s string
	_ = json.Unmarshal(body, &s)
	return s
}

func (r *recorder) call(body []byte) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls[name(body)]++
	return r.calls[name(body)]
}

func (r *recorder) count(body string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls[body]
}

// A failing event is retried after each delay with its original routing key,
// an event that keeps failing or can never succeed lands on the dead queue,
// and nothing is lost on the way.
func TestConsumerRetriesThenDeadLetters(t *testing.T) {
	url := startRabbitMQ(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	consumer := Consumer{
		URL: url, Queue: "test.events", RoutingKeys: []string{PaymentConfirmedType},
		RetryDelays: []time.Duration{300 * time.Millisecond, 300 * time.Millisecond}, Prefetch: 5, ReconnectDelay: time.Second,
	}
	seen := &recorder{calls: map[string]int{}}
	var wrongKeys sync.Map
	go consumer.Run(ctx, func(_ context.Context, routingKey string, body []byte) error {
		attempt := seen.call(body)
		if routingKey != PaymentConfirmedType {
			wrongKeys.Store(name(body), routingKey)
		}
		switch name(body) {
		case "flaky":
			if attempt < 3 { // succeeds on the third try, after two retries
				return errors.New("temporarily unavailable")
			}
		case "broken":
			return errors.New("always fails")
		case "contradiction":
			return fmt.Errorf("%w: terms differ", ErrPermanent)
		}
		return nil
	})

	publisher, err := NewRabbitMQPublisher(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = publisher.Close() })
	// The consumer declares the queue; wait until publishes are routable.
	deadline := time.Now().Add(20 * time.Second)
	for {
		if err = publisher.Publish(ctx, PaymentConfirmedType, "warmup"); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("queue never became routable: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
	}
	for _, body := range []string{"flaky", "broken", "contradiction"} {
		if err := publisher.Publish(ctx, PaymentConfirmedType, body); err != nil {
			t.Fatal(err)
		}
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	ch, err := conn.Channel()
	if err != nil {
		t.Fatal(err)
	}
	waitFor := func(what string, ok func() bool) {
		t.Helper()
		for end := time.Now().Add(15 * time.Second); !ok(); time.Sleep(100 * time.Millisecond) {
			if time.Now().After(end) {
				t.Fatalf("timed out waiting for %s", what)
			}
		}
	}
	deadMessages := func() int {
		queue, err := ch.QueueInspect(consumer.deadQueue())
		if err != nil {
			return -1
		}
		return queue.Messages
	}
	waitFor("flaky to succeed and the others to be dead-lettered", func() bool {
		return seen.count("flaky") == 3 && seen.count("broken") == 3 && seen.count("contradiction") == 1 && deadMessages() == 2
	})
	time.Sleep(time.Second) // a stray extra delivery would show up now
	if seen.count("flaky") != 3 || seen.count("broken") != 3 || seen.count("contradiction") != 1 || deadMessages() != 2 {
		t.Fatalf("flaky=%d broken=%d contradiction=%d dead=%d", seen.count("flaky"), seen.count("broken"), seen.count("contradiction"), deadMessages())
	}
	wrongKeys.Range(func(body, key any) bool {
		t.Errorf("%s was retried as %q, want %q", body, key, PaymentConfirmedType)
		return true
	})

	dead := map[string]string{}
	for range 2 {
		msg, ok, err := ch.Get(consumer.deadQueue(), true)
		if err != nil || !ok {
			t.Fatalf("read dead queue: %v %v", ok, err)
		}
		dead[name(msg.Body)] = fmt.Sprint(msg.Headers["x-failure-reason"])
	}
	if dead["broken"] != "always fails" || dead["contradiction"] != "permanent event failure: terms differ" {
		t.Fatalf("dead queue holds %v", dead)
	}
}
