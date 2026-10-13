package smtp

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	stdsmtp "net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/moomaideng/eventory/services/notification/internal/usecases"
)

type Config struct {
	Host, Port, Username, Password, FromAddress, FromName string
	StartTLS                                              bool
	Timeout                                               time.Duration
}

type Sender struct{ cfg Config }

func NewSender(cfg Config) *Sender { return &Sender{cfg: cfg} }

type SendError struct {
	err       error
	permanent bool
}

func (err *SendError) Error() string   { return err.err.Error() }
func (err *SendError) Unwrap() error   { return err.err }
func (err *SendError) Permanent() bool { return err.permanent }

func (sender *Sender) Send(ctx context.Context, message usecases.EmailMessage) error {
	dialer := net.Dialer{Timeout: sender.cfg.Timeout}
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(sender.cfg.Host, sender.cfg.Port))
	if err != nil {
		return classify(err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(sender.cfg.Timeout))

	client, err := stdsmtp.NewClient(conn, sender.cfg.Host)
	if err != nil {
		return classify(err)
	}
	defer client.Close()
	if sender.cfg.StartTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return &SendError{err: errors.New("SMTP server does not advertise STARTTLS"), permanent: true}
		}
		if err := client.StartTLS(&tls.Config{MinVersion: tls.VersionTLS12, ServerName: sender.cfg.Host}); err != nil {
			return classify(err)
		}
	}
	if sender.cfg.Username != "" {
		if err := client.Auth(stdsmtp.PlainAuth("", sender.cfg.Username, sender.cfg.Password, sender.cfg.Host)); err != nil {
			return classify(err)
		}
	}
	if err := client.Mail(sender.cfg.FromAddress); err != nil {
		return classify(err)
	}
	if err := client.Rcpt(message.To); err != nil {
		return classify(err)
	}
	wc, err := client.Data()
	if err != nil {
		return classify(err)
	}
	payload := buildMIME(sender.cfg, message)
	if _, err := wc.Write(payload); err != nil {
		_ = wc.Close()
		return classify(err)
	}
	if err := wc.Close(); err != nil {
		return classify(err)
	}
	// Closing DATA successfully means the relay accepted responsibility for the
	// message. A failed QUIT must not cause a duplicate email retry.
	_ = client.Quit()
	return nil
}

func buildMIME(cfg Config, message usecases.EmailMessage) []byte {
	boundary := "eventory-registration-boundary"
	var out bytes.Buffer
	fmt.Fprintf(&out, "From: %s <%s>\r\n", safeHeader(cfg.FromName), safeHeader(cfg.FromAddress))
	fmt.Fprintf(&out, "To: %s\r\n", message.To)
	fmt.Fprintf(&out, "Subject: %s\r\n", message.Subject)
	fmt.Fprintf(&out, "Message-ID: %s\r\n", message.MessageID)
	fmt.Fprint(&out, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&out, "Content-Type: multipart/alternative; boundary=%q\r\n\r\n", boundary)
	fmt.Fprintf(&out, "--%s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s\r\n", boundary, message.TextBody)
	fmt.Fprintf(&out, "--%s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s\r\n", boundary, message.HTMLBody)
	fmt.Fprintf(&out, "--%s--\r\n", boundary)
	return out.Bytes()
}

func safeHeader(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "\r", " "), "\n", " ")
}

func classify(err error) error {
	message := strings.TrimSpace(err.Error())
	code := 0
	if len(message) >= 3 {
		code, _ = strconv.Atoi(message[:3])
	}
	return &SendError{err: err, permanent: code >= 500 && code <= 599}
}
