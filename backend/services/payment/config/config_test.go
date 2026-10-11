package config

import "testing"

func TestServiceKeyConfiguration(t *testing.T) {
	t.Setenv("HTTP_PORT", "8083")
	t.Setenv("DB_DSN", "postgresql://unused/payment_db")
	t.Setenv("ENVIRONMENT", "production")
	for _, key := range []string{"short", "local-payment-service-key-not-for-production"} {
		t.Setenv("PAYMENT_SERVICE_KEY", key)
		if _, err := Load(); err == nil {
			t.Fatalf("Production accepted unsafe key %q", key)
		}
	}
	t.Setenv("PAYMENT_SERVICE_KEY", "test-only-private-key-32-characters-or-more")
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
}
