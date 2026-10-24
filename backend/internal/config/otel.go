package config

// OTel holds OpenTelemetry settings under their standard OTEL_* env names.
// An empty endpoint disables tracing.
type OTel struct {
	ExporterOTLPEndpoint string `mapstructure:"otel_exporter_otlp_endpoint"`
}
