package proto

import (
	"context"
	"crypto/subtle"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/payment/internal/repositories"
	"github.com/moomaideng/eventory/services/payment/internal/usecases"
	"github.com/moomaideng/eventory/services/payment/models"
	paymentv1 "github.com/moomaideng/eventory/services/payment/pkg/proto/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/gorm"
)

type PaymentUseCase interface {
	Create(context.Context, usecases.CreatePayment) (*models.Invoice, error)
	GetForService(context.Context, uuid.UUID) (*models.Invoice, error)
	RefundRegistration(context.Context, uuid.UUID, string) (*models.Invoice, error)
}

type PaymentServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	uc PaymentUseCase
}

func NewPaymentServer(uc PaymentUseCase) *PaymentServer { return &PaymentServer{uc: uc} }

// ServiceAuth fails closed, including when a server was constructed without a key.
func ServiceAuth(key string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		keys := md.Get("x-payment-service-key")
		if len(key) < 32 || len(keys) != 1 || subtle.ConstantTimeCompare([]byte(keys[0]), []byte(key)) != 1 {
			return nil, status.Error(codes.Unauthenticated, "Trusted service authentication required")
		}
		return handler(ctx, req)
	}
}

func (s *PaymentServer) CreatePayment(ctx context.Context, req *paymentv1.CreatePaymentRequest) (*paymentv1.CreatePaymentResponse, error) {
	payer, err := parseID(req.GetPayerAccountId())
	if err != nil {
		return nil, err
	}
	if req.GetExpiresAt() == nil || req.GetExpiresAt().CheckValid() != nil {
		return nil, status.Error(codes.InvalidArgument, "Invalid expiry timestamp")
	}
	invoice, err := s.uc.Create(ctx, usecases.CreatePayment{
		PayerAccountID: payer, Source: req.GetSource(), Reference: req.GetReference(), IdempotencyKey: req.GetIdempotencyKey(),
		Description: req.GetDescription(), AmountMinor: req.GetAmountMinor(), Currency: req.GetCurrency(), ExpiresAt: req.GetExpiresAt().AsTime(),
	})
	if err != nil {
		return nil, paymentError(err)
	}
	return &paymentv1.CreatePaymentResponse{Payment: toPayment(invoice)}, nil
}

func (s *PaymentServer) GetPayment(ctx context.Context, req *paymentv1.GetPaymentRequest) (*paymentv1.GetPaymentResponse, error) {
	id, err := parseID(req.GetId())
	if err != nil {
		return nil, err
	}
	invoice, err := s.uc.GetForService(ctx, id)
	if err != nil {
		return nil, paymentError(err)
	}
	return &paymentv1.GetPaymentResponse{Payment: toPayment(invoice)}, nil
}

func (s *PaymentServer) RefundRegistration(ctx context.Context, req *paymentv1.RefundRegistrationRequest) (*paymentv1.RefundRegistrationResponse, error) {
	id, err := parseID(req.GetId())
	if err != nil {
		return nil, err
	}
	entry := strings.TrimSpace(req.GetEntryId())
	if entry == "" {
		return nil, status.Error(codes.InvalidArgument, "Invalid entry ID")
	}
	invoice, err := s.uc.RefundRegistration(ctx, id, entry)
	if err != nil {
		return nil, paymentError(err)
	}
	return &paymentv1.RefundRegistrationResponse{Payment: toPayment(invoice)}, nil
}

func parseID(value string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, status.Error(codes.InvalidArgument, "Invalid UUID")
	}
	return id, nil
}

func paymentError(err error) error {
	switch {
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "Request canceled")
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "Request deadline exceeded")
	case errors.Is(err, usecases.ErrInvalid):
		return status.Error(codes.InvalidArgument, "Invalid payment input")
	case errors.Is(err, repositories.ErrConflict):
		return status.Error(codes.FailedPrecondition, "Invoice conflicts with the request or its current state")
	case errors.Is(err, gorm.ErrRecordNotFound):
		return status.Error(codes.NotFound, "Invoice not found")
	default:
		return status.Error(codes.Internal, "Unable to process payment")
	}
}

func toPayment(p *models.Invoice) *paymentv1.Payment {
	out := &paymentv1.Payment{Id: p.ID.String(), PayerAccountId: p.PayerAccountID.String(), Source: p.Source, Reference: p.Reference,
		Description: p.Description, AmountMinor: p.AmountMinor, Currency: p.Currency, Provider: p.Provider, Status: string(p.Status(time.Now())),
		CreatedAt: timestamppb.New(p.CreatedAt), ExpiresAt: timestamppb.New(p.ExpiresAt)}
	if p.PaidAt != nil {
		out.PaidAt = timestamppb.New(*p.PaidAt)
	}
	if p.RefundedAt != nil {
		out.RefundedAt = timestamppb.New(*p.RefundedAt)
	}
	return out
}
