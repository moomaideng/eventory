package money

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	// DefaultCurrency is the project's primary settlement and display currency.
	DefaultCurrency = "THB"

	// MaxMinor is the maximum minor-unit amount representable losslessly in JSON/JavaScript (Number.MAX_SAFE_INTEGER).
	MaxMinor int64 = 9007199254740991
)

var (
	// ErrInvalidCurrency indicates a blank or unsupported ISO-4217 currency code.
	ErrInvalidCurrency = errors.New("unsupported currency")

	// ErrInvalidAmount indicates a negative or out-of-range minor-unit amount.
	ErrInvalidAmount = errors.New("invalid money amount")
)

type currencySpec struct {
	minorUnit uint64
	decimals  int
}

var supportedCurrencies = map[string]currencySpec{
	DefaultCurrency: {minorUnit: 100, decimals: 2},
}

// Money is an exact monetary value stored in the currency's smallest minor unit
// (for example, satang for THB).
type Money struct {
	AmountMinor int64
	Currency    string
}

// CurrencyOrDefault returns a trimmed uppercase currency code, falling back to DefaultCurrency when blank.
func CurrencyOrDefault(currency string) string {
	if normalized := normalizeCurrency(currency); normalized != "" {
		return normalized
	}
	return DefaultCurrency
}

// FromMinor validates a non-negative minor-unit amount and supported currency.
// Callers that forbid zero (such as invoices or tier prices) should also check amt.AmountMinor == 0.
func FromMinor(amountMinor int64, currency string) (Money, error) {
	normalized := normalizeCurrency(currency)
	if _, ok := supportedCurrencies[normalized]; !ok {
		return Money{}, ErrInvalidCurrency
	}
	if amountMinor < 0 || amountMinor > MaxMinor {
		return Money{}, ErrInvalidAmount
	}
	return Money{AmountMinor: amountMinor, Currency: normalized}, nil
}

// Format renders a minor-unit amount with its currency prefix using exact integer arithmetic
// (for example, "THB 1,250.00").
func Format(amountMinor int64, currency string) string {
	resolvedCurrency := CurrencyOrDefault(currency)
	spec, ok := supportedCurrencies[resolvedCurrency]
	if !ok {
		spec = supportedCurrencies[DefaultCurrency]
	}

	negative := amountMinor < 0
	var abs uint64
	if negative {
		abs = uint64(-amountMinor)
	} else {
		abs = uint64(amountMinor)
	}

	whole := strconv.FormatUint(abs/spec.minorUnit, 10)
	for i := len(whole) - 3; i > 0; i -= 3 {
		whole = whole[:i] + "," + whole[i:]
	}
	if negative {
		whole = "-" + whole
	}
	if spec.decimals <= 0 {
		return fmt.Sprintf("%s %s", resolvedCurrency, whole)
	}

	frac := strconv.FormatUint(abs%spec.minorUnit, 10)
	if len(frac) < spec.decimals {
		frac = strings.Repeat("0", spec.decimals-len(frac)) + frac
	}
	return fmt.Sprintf("%s %s.%s", resolvedCurrency, whole, frac)
}

func normalizeCurrency(currency string) string {
	return strings.ToUpper(strings.TrimSpace(currency))
}
