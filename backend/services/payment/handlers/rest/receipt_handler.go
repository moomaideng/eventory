package rest

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/internal/middlewares"
	"github.com/moomaideng/eventory/services/payment/internal/receiptpdf"
	"github.com/moomaideng/eventory/services/payment/internal/usecases"
	"github.com/moomaideng/eventory/services/payment/models"
	"net/http"
	"time"
)

type ReceiptResponse struct {
	ID             uuid.UUID `json:"id"`
	InvoiceID      uuid.UUID `json:"invoiceId"`
	PayerAccountID uuid.UUID `json:"payerAccountId"`
	Source         string    `json:"source"`
	Reference      string    `json:"reference"`
	Description    string    `json:"description"`
	AmountMinor    int64     `json:"amountMinor"`
	Currency       string    `json:"currency" enum:"THB"`
	Provider       string    `json:"provider" enum:"MOCK"`
	PaidAt         time.Time `json:"paidAt"`
}

func receiptResponse(invoice *models.Invoice) ReceiptResponse {
	return ReceiptResponse{ID: invoice.Receipt.ID, InvoiceID: invoice.ID, PayerAccountID: invoice.PayerAccountID,
		Source: invoice.Source, Reference: invoice.Reference, Description: invoice.Description,
		AmountMinor: invoice.AmountMinor, Currency: invoice.Currency, Provider: invoice.Provider, PaidAt: *invoice.PaidAt}
}

func registerReceiptRoutes(api huma.API, uc *usecases.PaymentUseCase) {
	security := []map[string][]string{{"bearer": {}}}
	huma.Register(api, huma.Operation{OperationID: "get-my-receipt", Method: http.MethodGet, Path: "/{id}/receipt", Summary: "View the receipt for an owned paid invoice", Security: security},
		func(ctx context.Context, input *PaymentIDInput) (*struct{ Body ReceiptResponse }, error) {
			payer, err := middlewares.GetAuthUserID(ctx)
			if err != nil {
				return nil, huma.Error401Unauthorized("Authentication required")
			}
			invoice, err := uc.GetReceipt(ctx, payer, input.ID)
			if err != nil {
				return nil, paymentError(err)
			}
			return &struct{ Body ReceiptResponse }{receiptResponse(invoice)}, nil
		})
	huma.Register(api, huma.Operation{OperationID: "download-my-receipt", Method: http.MethodGet, Path: "/{id}/receipt.pdf", Summary: "Download a mock receipt as PDF", Security: security, Responses: map[string]*huma.Response{
		"200": {Description: "Mock receipt PDF", Content: map[string]*huma.MediaType{"application/pdf": {Schema: &huma.Schema{Type: "string", Format: "binary"}}}},
	}}, func(ctx context.Context, input *PaymentIDInput) (*struct {
		ContentType        string `header:"Content-Type"`
		ContentDisposition string `header:"Content-Disposition"`
		Body               []byte
	}, error) {
		payer, err := middlewares.GetAuthUserID(ctx)
		if err != nil {
			return nil, huma.Error401Unauthorized("Authentication required")
		}
		invoice, err := uc.GetReceipt(ctx, payer, input.ID)
		if err != nil {
			return nil, paymentError(err)
		}
		pdf, err := receiptpdf.Generate(invoice)
		if err != nil {
			return nil, paymentError(err)
		}
		return &struct {
			ContentType        string `header:"Content-Type"`
			ContentDisposition string `header:"Content-Disposition"`
			Body               []byte
		}{"application/pdf", `attachment; filename="receipt-` + invoice.Receipt.ID.String() + `.pdf"`, pdf}, nil
	})
}
