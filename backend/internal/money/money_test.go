package money

import (
	"errors"
	"testing"
)

func TestCurrencyOrDefault(t *testing.T) {
	if got := CurrencyOrDefault("   "); got != DefaultCurrency {
		t.Fatalf("CurrencyOrDefault(blank) = %q, want %q", got, DefaultCurrency)
	}
	if got := CurrencyOrDefault(" thb "); got != "THB" {
		t.Fatalf("CurrencyOrDefault(thb) = %q, want THB", got)
	}
}

func TestFromMinor(t *testing.T) {
	m, err := FromMinor(0, "thb")
	if err != nil {
		t.Fatalf("FromMinor(0) unexpected error: %v", err)
	}
	if m.AmountMinor != 0 || m.Currency != "THB" {
		t.Fatalf("unexpected zero Money: %+v", m)
	}

	m, err = FromMinor(MaxMinor, "THB")
	if err != nil {
		t.Fatalf("FromMinor(MaxMinor) unexpected error: %v", err)
	}
	if m.AmountMinor != MaxMinor || m.Currency != "THB" {
		t.Fatalf("unexpected MaxMinor Money: %+v", m)
	}

	for _, tc := range []struct {
		name     string
		amount   int64
		currency string
		wantErr  error
	}{
		{name: "negative minor", amount: -1, currency: "THB", wantErr: ErrInvalidAmount},
		{name: "above max minor", amount: MaxMinor + 1, currency: "THB", wantErr: ErrInvalidAmount},
		{name: "unsupported currency", amount: 100, currency: "USD", wantErr: ErrInvalidCurrency},
		{name: "blank currency", amount: 100, currency: "", wantErr: ErrInvalidCurrency},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := FromMinor(tc.amount, tc.currency); !errors.Is(err, tc.wantErr) {
				t.Fatalf("got err %v, want %v", err, tc.wantErr)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name     string
		minor    int64
		currency string
		want     string
	}{
		{name: "zero", minor: 0, currency: "THB", want: "THB 0.00"},
		{name: "sub-unit", minor: 5, currency: "THB", want: "THB 0.05"},
		{name: "standard", minor: 125000, currency: "thb", want: "THB 1,250.00"},
		{name: "blank defaults to THB", minor: 125050, currency: "", want: "THB 1,250.50"},
		{name: "negative", minor: -125005, currency: "THB", want: "THB -1,250.05"},
		{name: "max safe minor exact", minor: MaxMinor, currency: "THB", want: "THB 90,071,992,547,409.91"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := Format(tc.minor, tc.currency); got != tc.want {
				t.Fatalf("Format = %q, want %q", got, tc.want)
			}
		})
	}
}
