package paymenthttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Client struct {
	url, key string
	http     *http.Client
}

func New(url, key string) *Client {
	return &Client{
		url: strings.TrimRight(url, "/"),
		key: key,
		http: &http.Client{
			Timeout:   5 * time.Second,
			Transport: otelhttp.NewTransport(http.DefaultTransport),
		},
	}
}

// Creation is idempotent and also returns the latest status on retries.
func (c *Client) EnsureInvoice(ctx context.Context, input ports.RegistrationPaymentRequest) (*ports.RegistrationPayment, error) {
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/internal/v1/payments", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Payment-Service-Key", c.key)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("payment service returned %d", res.StatusCode)
	}
	var payment ports.RegistrationPayment
	if err := json.NewDecoder(io.LimitReader(res.Body, 65536)).Decode(&payment); err != nil {
		return nil, err
	}
	return &payment, nil
}

func (c *Client) RefundRegistration(ctx context.Context, paymentID, entryID uuid.UUID) (*ports.RegistrationPayment, error) {
	body, err := json.Marshal(struct {
		EntryID uuid.UUID `json:"entryId"`
	}{entryID})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/internal/v1/payments/"+paymentID.String()+"/refund", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Payment-Service-Key", c.key)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payment service returned %d", res.StatusCode)
	}
	var payment ports.RegistrationPayment
	if err := json.NewDecoder(io.LimitReader(res.Body, 65536)).Decode(&payment); err != nil {
		return nil, err
	}
	if payment.ID != paymentID || payment.Status != "REFUNDED" || payment.Source != "TEAM_REGISTRATION" || !strings.HasPrefix(payment.Reference, entryID.String()+":") {
		return nil, fmt.Errorf("payment service returned a mismatched refund")
	}
	return &payment, nil
}
