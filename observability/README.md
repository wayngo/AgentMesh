# Observability

The gateway instruments HTTP requests with OpenTelemetry (`agentmesh.gateway`) and exposes request totals, error totals, and cumulative latency at `/debug/metrics` using the standard expvar format. Structured request logs are emitted with `slog`.

OpenTelemetry exporter configuration can be added through the standard OTEL environment variables when a collector is available; the gateway remains functional without one.