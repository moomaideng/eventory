package paymentgrpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/grpcx"
	paymentv1 "github.com/moomaideng/eventory/services/payment/pkg/proto/payment/v1"
	"github.com/moomaideng/eventory/services/tournament/internal/ports"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Client struct {
	api  paymentv1.PaymentServiceClient
	key  string
	conn *grpc.ClientConn
}

var _ ports.PaymentService = (*Client)(nil)
var _ ports.RegistrationRefundService = (*Client)(nil)

func New(addr, key string, slowRequestMS int) (*Client, error) {
	conn, err := grpcx.NewClient(addr, slowRequestMS, grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	if err != nil {
		return nil, fmt.Errorf("dial payment gRPC: %w", err)
	}
	return &Client{api: paymentv1.NewPaymentServiceClient(conn), key: key, conn: conn}, nil
}

func (c *Client) Close() error { return c.conn.Close() }

func (c *Client) callContext(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	md, _ := metadata.FromOutgoingContext(ctx)
	md = md.Copy()
	md.Set("x-payment-service-key", c.key)
	return metadata.NewOutgoingContext(ctx, md), cancel
}

// Both operations are idempotent, so grpcx's bounded transient retries are safe.
func (c *Client) EnsureInvoice(ctx context.Context, in ports.RegistrationPaymentRequest) (*ports.RegistrationPayment, error) {
	ctx, cancel := c.callContext(ctx)
	defer cancel()
	res, err := c.api.CreatePayment(ctx, &paymentv1.CreatePaymentRequest{PayerAccountId: in.PayerAccountID.String(), Source: in.Source,
		Reference: in.Reference, IdempotencyKey: in.IdempotencyKey, Description: in.Description, AmountMinor: in.AmountMinor,
		Currency: in.Currency, ExpiresAt: timestamppb.New(in.ExpiresAt)})
	if err != nil {
		return nil, fmt.Errorf("create payment: %w", err)
	}
	return fromPayment(res.GetPayment())
}

func (c *Client) RefundRegistration(ctx context.Context, paymentID, entryID uuid.UUID) (*ports.RegistrationPayment, error) {
	ctx, cancel := c.callContext(ctx)
	defer cancel()
	res, err := c.api.RefundRegistration(ctx, &paymentv1.RefundRegistrationRequest{Id: paymentID.String(), EntryId: entryID.String()})
	if err != nil {
		return nil, fmt.Errorf("refund registration: %w", err)
	}
	p, err := fromPayment(res.GetPayment())
	if err != nil {
		return nil, err
	}
	if p.ID != paymentID || p.Status != "REFUNDED" || p.Source != "TEAM_REGISTRATION" || !strings.HasPrefix(p.Reference, entryID.String()+":") {
		return nil, fmt.Errorf("payment service returned a mismatched refund")
	}
	return p, nil
}

func fromPayment(p *paymentv1.Payment) (*ports.RegistrationPayment, error) {
	if p == nil {
		return nil, fmt.Errorf("payment service returned no payment")
	}
	id, err := uuid.Parse(p.GetId())
	if err != nil || id == uuid.Nil {
		return nil, fmt.Errorf("invalid payment ID in response")
	}
	payer, err := uuid.Parse(p.GetPayerAccountId())
	if err != nil || payer == uuid.Nil {
		return nil, fmt.Errorf("invalid payer ID in response")
	}
	if p.GetExpiresAt() == nil || p.GetExpiresAt().CheckValid() != nil {
		return nil, fmt.Errorf("invalid expiry in response")
	}
	return &ports.RegistrationPayment{ID: id, PayerAccountID: payer, Source: p.GetSource(), Reference: p.GetReference(), AmountMinor: p.GetAmountMinor(), Currency: p.GetCurrency(), ExpiresAt: p.GetExpiresAt().AsTime(), Status: p.GetStatus()}, nil
}
