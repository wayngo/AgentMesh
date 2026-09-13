# Phase plan

Phase 0: scaffold. Phase 1: Go REST gateway to Python worker over gRPC using the versioned protobuf wire contract. Phase 2: PostgreSQL. Phase 3: Temporal. Later phases add providers, WebSockets, Redis, Kafka, auth, telemetry, containers, Kubernetes, AWS, GitOps, Envoy, load and failure testing.

## Phase 1 completion criteria

- The worker serves `RunWorker/ExecuteRun` on gRPC port `9000`.
- The gateway calls the worker for `POST /v1/runs` and returns the worker result.
- Worker failures are persisted as failed runs and returned as HTTP 502.
- Go and Python tests cover successful execution, failure behavior, protocol round-trips, malformed payloads, and worker startup.

Phase 2 is complete: run state can persist in PostgreSQL, with memory storage retained as the default when DATABASE_URL is absent. Phase 3 is complete: when TEMPORAL_ADDRESS is configured, runs are started as durable Temporal workflows, activities call the gRPC worker with retries, and completion is persisted. Phase 4 is complete: the worker has a provider abstraction with mock, OpenAI, and Anthropic adapters selected by AGENTMESH_LLM_PROVIDER, with provider calls isolated behind tests. Phase 5 is complete: the Vite-powered frontend can submit runs, poll workflow status, render results, and display failures. Phase 6 is complete: the gateway streams run updates over WebSockets and the frontend consumes them with polling fallback. Phase 7 is complete: Redis can wrap the source-of-truth run store with write-through caching and cache-miss fallback, enabled with REDIS_URL. Phase 8 is complete: run updates can be published as versioned Kafka events through the optional KAFKA_BROKERS/KAFKA_TOPIC configuration. Phase 9 is complete: optional HS256 JWT authentication protects gateway run endpoints while keeping health checks public, enabled with AUTH_JWT_SECRET. The project is ready for Phase 10, observability.