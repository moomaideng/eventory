package models

import (
	"testing"
	"time"
)

func TestInvoiceStatus(t *testing.T) {
	deadline := time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)
	invoice := Invoice{ExpiresAt: deadline}
	for _, test := range []struct {
		now  time.Time
		want Status
	}{
		{deadline.Add(-time.Nanosecond), Unpaid}, {deadline, Expired}, {deadline.Add(time.Hour), Expired},
	} {
		if got := invoice.Status(test.now); got != test.want {
			t.Fatalf("got %s, want %s", got, test.want)
		}
	}
	paidAt := deadline.Add(-time.Minute)
	invoice.PaidAt = &paidAt
	if invoice.Status(deadline.Add(time.Hour)) != Paid {
		t.Fatal("paid invoice must stay paid after its deadline")
	}
}
