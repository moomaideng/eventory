package payments_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/payment/handlers/rest"
	"github.com/moomaideng/eventory/services/payment/models"
	paymentv1 "github.com/moomaideng/eventory/services/payment/pkg/proto/payment/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func paymentClient(t *testing.T, server *grpc.Server) paymentv1.PaymentServiceClient {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = server.Serve(lis) }()
	t.Cleanup(server.Stop)
	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return paymentv1.NewPaymentServiceClient(conn)
}

// Keep existing database/concurrency assertions while exercising trusted operations
// through real gRPC. The recorder only normalizes results for the shared assertions.
func internalRPC(client paymentv1.PaymentServiceClient, method, path, body, key string) *httptest.ResponseRecorder {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "x-payment-service-key", key)
	var payment *paymentv1.Payment
	var err error
	switch {
	case path == "/internal/v1/payments":
		var in struct {
			PayerAccountID, Source, Reference, IdempotencyKey, Description, Currency string
			AmountMinor                                                              int64
			ExpiresAt                                                                time.Time
		}
		err = json.Unmarshal([]byte(body), &in)
		if err == nil {
			var res *paymentv1.CreatePaymentResponse
			res, err = client.CreatePayment(ctx, &paymentv1.CreatePaymentRequest{PayerAccountId: in.PayerAccountID, Source: in.Source, Reference: in.Reference,
				IdempotencyKey: in.IdempotencyKey, Description: in.Description, AmountMinor: in.AmountMinor, Currency: in.Currency, ExpiresAt: timestamppb.New(in.ExpiresAt)})
			payment = res.GetPayment()
		}
	case strings.HasSuffix(path, "/refund"):
		var in struct{ EntryID string }
		err = json.Unmarshal([]byte(body), &in)
		if err == nil {
			var res *paymentv1.RefundRegistrationResponse
			res, err = client.RefundRegistration(ctx, &paymentv1.RefundRegistrationRequest{Id: strings.Split(path, "/")[4], EntryId: in.EntryID})
			payment = res.GetPayment()
		}
	default:
		var res *paymentv1.GetPaymentResponse
		res, err = client.GetPayment(ctx, &paymentv1.GetPaymentRequest{Id: strings.TrimPrefix(path, "/internal/v1/payments/")})
		payment = res.GetPayment()
	}
	w := httptest.NewRecorder()
	if err != nil {
		code := map[codes.Code]int{codes.Unauthenticated: 401, codes.InvalidArgument: 422, codes.NotFound: 404, codes.FailedPrecondition: 409}[status.Code(err)]
		if code == 0 {
			code = 500
		}
		w.WriteHeader(code)
		_, _ = w.WriteString(err.Error())
		return w
	}
	out := rest.PaymentResponse{ID: uuid.MustParse(payment.Id), PayerAccountID: uuid.MustParse(payment.PayerAccountId), Source: payment.Source,
		Reference: payment.Reference, Description: payment.Description, AmountMinor: payment.AmountMinor, Currency: payment.Currency,
		Provider: payment.Provider, Status: models.Status(payment.Status), CreatedAt: payment.CreatedAt.AsTime(), ExpiresAt: payment.ExpiresAt.AsTime()}
	if payment.PaidAt != nil {
		v := payment.PaidAt.AsTime()
		out.PaidAt = &v
	}
	if payment.RefundedAt != nil {
		v := payment.RefundedAt.AsTime()
		out.RefundedAt = &v
	}
	_ = json.NewEncoder(w).Encode(out)
	return w
}
