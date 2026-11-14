package proto

import (
	"context"
	"errors"
	"testing"
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

type stubPayments struct {
	err   error
	calls int
}

func (s *stubPayments) Create(context.Context, usecases.CreatePayment) (*models.Invoice, error) {
	s.calls++
	return nil, s.err
}
func (s *stubPayments) GetForService(context.Context, uuid.UUID) (*models.Invoice, error) {
	s.calls++
	return nil, s.err
}
func (s *stubPayments) RefundRegistration(context.Context, uuid.UUID, string) (*models.Invoice, error) {
	s.calls++
	return nil, s.err
}

func TestServiceAuthentication(t *testing.T) {
	key := "test-payment-service-key-at-least-32-characters"
	for _, tc := range []struct {
		name, configured string
		keys             []string
		want             codes.Code
	}{
		{"valid", key, []string{key}, codes.OK},
		{"missing", key, nil, codes.Unauthenticated},
		{"wrong", key, []string{"wrong"}, codes.Unauthenticated},
		{"duplicate", key, []string{key, key}, codes.Unauthenticated},
		{"unconfigured", "", []string{key}, codes.Unauthenticated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{"x-payment-service-key": tc.keys})
			called := false
			_, err := ServiceAuth(tc.configured)(ctx, nil, &grpc.UnaryServerInfo{}, func(context.Context, any) (any, error) { called = true; return nil, nil })
			if status.Code(err) != tc.want || called != (tc.want == codes.OK) {
				t.Fatalf("called=%v, err=%v", called, err)
			}
		})
	}
}

func TestRPCValidationAndErrors(t *testing.T) {
	stub := &stubPayments{}
	server := NewPaymentServer(stub)
	ctx := context.Background()
	for _, id := range []string{"bad", uuid.Nil.String()} {
		if _, err := server.GetPayment(ctx, &paymentv1.GetPaymentRequest{Id: id}); status.Code(err) != codes.InvalidArgument {
			t.Fatal(err)
		}
	}
	for _, entryID := range []string{"", "   "} {
		if _, err := server.RefundRegistration(ctx, &paymentv1.RefundRegistrationRequest{Id: uuid.NewString(), EntryId: entryID}); status.Code(err) != codes.InvalidArgument {
			t.Fatal(err)
		}
	}
	for _, expiry := range []*timestamppb.Timestamp{nil, {Seconds: 253402300800}} {
		if _, err := server.CreatePayment(ctx, &paymentv1.CreatePaymentRequest{PayerAccountId: uuid.NewString(), ExpiresAt: expiry}); status.Code(err) != codes.InvalidArgument {
			t.Fatal(err)
		}
	}
	if stub.calls != 0 {
		t.Fatal("invalid requests reached domain")
	}
	for _, tc := range []struct {
		err  error
		want codes.Code
	}{
		{usecases.ErrInvalid, codes.InvalidArgument}, {repositories.ErrConflict, codes.FailedPrecondition},
		{gorm.ErrRecordNotFound, codes.NotFound}, {context.Canceled, codes.Canceled},
		{context.DeadlineExceeded, codes.DeadlineExceeded}, {errors.New("private database details"), codes.Internal},
	} {
		stub.err = tc.err
		_, err := server.CreatePayment(ctx, &paymentv1.CreatePaymentRequest{PayerAccountId: uuid.NewString(), ExpiresAt: timestamppb.New(time.Now().Add(time.Hour))})
		if status.Code(err) != tc.want {
			t.Fatalf("got %v, want %v", err, tc.want)
		}
		if tc.want == codes.Internal && status.Convert(err).Message() != "Unable to process payment" {
			t.Fatal("leaked internal error")
		}
	}
}
