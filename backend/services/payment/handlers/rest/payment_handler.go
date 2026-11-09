package rest

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/services/payment/internal/repositories"
	"github.com/moomaideng/eventory/services/payment/internal/usecases"
	"github.com/moomaideng/eventory/services/payment/models"
	"gorm.io/gorm"
)

type PaymentResponse struct {
	PayerAccountID uuid.UUID     `json:"payerAccountId"`
	ID             uuid.UUID     `json:"id"`
	Source         string        `json:"source"`
	Reference      string        `json:"reference"`
	Description    string        `json:"description"`
	AmountMinor    int64         `json:"amountMinor" doc:"Amount in satang"`
	Currency       string        `json:"currency" enum:"THB"`
	Provider       string        `json:"provider" enum:"MOCK"`
	Status         models.Status `json:"status" enum:"UNPAID,PAID,REFUNDED,EXPIRED,CANCELLED"`
	CreatedAt      time.Time     `json:"createdAt"`
	ExpiresAt      time.Time     `json:"expiresAt"`
	PaidAt         *time.Time    `json:"paidAt"`
	RefundedAt     *time.Time    `json:"refundedAt"`
	InvoicePath    string        `json:"invoicePath"`
	CheckoutPath   string        `json:"checkoutPath"`
}

type PaymentOutput struct{ Body PaymentResponse }
type PaymentIDInput struct {
	ID uuid.UUID `path:"id"`
}

func response(payment *models.Invoice) PaymentResponse {
	return PaymentResponse{PayerAccountID: payment.PayerAccountID, ID: payment.ID, Source: payment.Source, Reference: payment.Reference, Description: payment.Description,
		AmountMinor: payment.AmountMinor, Currency: payment.Currency, Provider: payment.Provider, Status: payment.Status(time.Now()),
		CreatedAt: payment.CreatedAt, ExpiresAt: payment.ExpiresAt, PaidAt: payment.PaidAt, RefundedAt: payment.RefundedAt,
		InvoicePath: "/payments/" + payment.ID.String(), CheckoutPath: "/payments/" + payment.ID.String() + "/checkout"}
}

func paymentError(err error) error {
	switch {
	case errors.Is(err, usecases.ErrInvalid):
		return huma.Error422UnprocessableEntity("Invalid payment input or date range")
	case errors.Is(err, repositories.ErrConflict):
		return huma.Error409Conflict("Invoice conflicts with an existing request, is already paid, or has expired")
	case errors.Is(err, gorm.ErrRecordNotFound):
		return huma.Error404NotFound("Invoice or receipt not found")
	default:
		return huma.Error500InternalServerError("Unable to process payment", err)
	}
}

func RegisterPaymentRoutes(api huma.API, uc *usecases.PaymentUseCase) {
	registerReceiptRoutes(api, uc)
	security := []map[string][]string{{"bearer": {}}}
	huma.Register(api, huma.Operation{OperationID: "get-my-payment", Method: http.MethodGet, Path: "/{id}", Summary: "View an owned payment", Security: security},
		func(ctx context.Context, input *PaymentIDInput) (*PaymentOutput, error) {
			payer, err := middlewares.GetAuthUserID(ctx)
			if err != nil {
				return nil, huma.Error401Unauthorized("Authentication required")
			}
			payment, err := uc.GetOwned(ctx, payer, input.ID)
			if err != nil {
				return nil, paymentError(err)
			}
			return &PaymentOutput{Body: response(payment)}, nil
		})
	huma.Register(api, huma.Operation{OperationID: "list-my-payments", Method: http.MethodGet, Path: "", Summary: "Personal invoices, newest first; date filters use UTC", Security: security},
		func(ctx context.Context, input *struct {
			From  string `query:"from" doc:"Inclusive creation date YYYY-MM-DD (UTC)"`
			To    string `query:"to" doc:"Inclusive creation date YYYY-MM-DD (UTC)"`
			Page  int    `query:"page" default:"1" minimum:"1" maximum:"1000000"`
			Limit int    `query:"limit" default:"20" minimum:"1" maximum:"100"`
		}) (*struct {
			Body struct {
				Items []PaymentResponse `json:"items"`
				Total int64             `json:"total"`
			}
		}, error) {
			payer, err := middlewares.GetAuthUserID(ctx)
			if err != nil {
				return nil, huma.Error401Unauthorized("Authentication required")
			}
			payments, total, err := uc.ListOwned(ctx, payer, input.From, input.To, input.Page, input.Limit)
			if err != nil {
				return nil, paymentError(err)
			}
			out := &struct {
				Body struct {
					Items []PaymentResponse `json:"items"`
					Total int64             `json:"total"`
				}
			}{}
			out.Body.Items = make([]PaymentResponse, 0, len(payments))
			for i := range payments {
				out.Body.Items = append(out.Body.Items, response(&payments[i]))
			}
			out.Body.Total = total
			return out, nil
		})
	huma.Register(api, huma.Operation{OperationID: "complete-mock-payment", Method: http.MethodPost, Path: "/{id}/mock-result", Summary: "Simulate a payment result; no money moves", Security: security},
		func(ctx context.Context, input *struct {
			ID   uuid.UUID `path:"id"`
			Body struct {
				Result models.MockResult `json:"result" enum:"SUCCEEDED,FAILED,CANCELLED"`
			}
		}) (*PaymentOutput, error) {
			payer, err := middlewares.GetAuthUserID(ctx)
			if err != nil {
				return nil, huma.Error401Unauthorized("Authentication required")
			}
			email, _ := middlewares.GetAuthEmail(ctx)
			payment, err := uc.CompleteMock(ctx, payer, input.ID, input.Body.Result, email)
			if err != nil {
				return nil, paymentError(err)
			}
			return &PaymentOutput{Body: response(payment)}, nil
		})
}
