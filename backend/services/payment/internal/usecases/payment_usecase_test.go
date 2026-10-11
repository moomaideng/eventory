package usecases

import "testing"

func TestHistoryFilter(t *testing.T) {
	for _, tc := range []struct {
		from, to    string
		page, limit int
		valid       bool
	}{
		{"", "", 1, 20, true}, {"2024-02-29", "2024-02-29", 1, 20, true},
		{"2025-02-29", "", 1, 20, false}, {"2026-01-02", "2026-01-01", 1, 20, false},
		{"2026-1-01", "", 1, 20, false}, {"", "bad", 1, 20, false},
		{"", "", 0, 20, false}, {"", "", 1, 101, false}, {"", "", 1000001, 20, false},
	} {
		filter, err := HistoryFilter(tc.from, tc.to, tc.page, tc.limit)
		if (err == nil) != tc.valid {
			t.Errorf("%+v: %v", tc, err)
		}
		if tc.from != "" && tc.from == tc.to && tc.valid && filter.Until.Sub(*filter.From).Hours() != 24 {
			t.Error("End date must include the entire day")
		}
	}
}
