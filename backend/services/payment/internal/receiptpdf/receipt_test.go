package receiptpdf

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/moomaideng/eventory/services/payment/models"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	invoice := models.Invoice{}
	if _, err := Generate(&invoice); err == nil {
		t.Fatal("unpaid invoice generated a receipt")
	}
	now := time.Now()
	invoice = models.Invoice{ID: uuid.New(), PayerAccountID: uuid.New(), Description: strings.Repeat("ค่าลงทะเบียน ", 20),
		Source: "test", Reference: strings.Repeat("x", 128), AmountMinor: 9007199254740991, Currency: "THB", Provider: "MOCK", PaidAt: &now,
		Receipt: &models.Receipt{ID: uuid.New()}}
	pdf, err := Generate(&invoice)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(pdf, []byte("%PDF-")) || !bytes.Contains(pdf, []byte("%%EOF")) {
		t.Fatal("invalid PDF document")
	}
	invoice.Description, invoice.Reference, invoice.Source = strings.Repeat("W", 255), strings.Repeat("W", 128), strings.Repeat("W", 64)
	pdf, err = Generate(&invoice)
	if err != nil {
		t.Fatal(err)
	}
	if pages := len(regexp.MustCompile(`/Type\s*/Page\b`).FindAll(pdf, -1)); pages != 1 {
		t.Fatalf("expected one A4 page, got %d", pages)
	}
}

func TestFormatAmount(t *testing.T) {
	for _, test := range []struct {
		minor int64
		want  string
	}{
		{1, "THB 0.01"}, {15001, "THB 150.01"}, {123456, "THB 1,234.56"}, {9007199254740991, "THB 90,071,992,547,409.91"},
	} {
		got := formatAmount(&models.Invoice{AmountMinor: test.minor, Currency: "THB"})
		if got != test.want {
			t.Fatalf("amount %d: got %s, want %s", test.minor, got, test.want)
		}
	}
}
