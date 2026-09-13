# AgentMesh

AgentMesh is a distributed AI-agent execution platform. This repository is being built in incremental phases. The current implementation includes the Go gateway, Python worker, frontend, persistence, orchestration, events, security, observability, and deployment layers through Phase 16.

## Current capabilities

- REST API: create and retrieve runs with `POST /v1/runs` and `GET /v1/runs/{id}`.
- gRPC: gateway-to-worker execution using the versioned contract in `proto/agentmesh/v1/run.proto`.
- Persistence: in-memory default or PostgreSQL through `DATABASE_URL`.
- Temporal: durable asynchronous workflows through `TEMPORAL_ADDRESS`.
- LLM providers: mock, OpenAI, and Anthropic through `AGENTMESH_LLM_PROVIDER`.
- Real-time updates: WebSocket endpoint `/v1/runs/{id}/events` with frontend polling fallback.
- Redis: optional write-through run cache through `REDIS_URL`.
- Kafka: optional versioned `run.updated` events through `KAFKA_BROKERS`.
- Security: optional HS256 JWT authentication through `AUTH_JWT_SECRET`.
- Observability: OpenTelemetry HTTP instrumentation, structured logs, and `/debug/metrics`.
- Deployment: Docker Compose, Kubernetes/Kustomize, and Helm.
- AWS foundation: Terraform-managed VPC, EKS, PostgreSQL, Redis, and Kafka.
- GitOps and edge routing: Argo CD application definitions and Envoy WebSocket-aware routing.
- Validation: k6 load tests, API failure checks, and production-readiness checklist.

## Local development

Start the Python worker:

```powershell
python -m pip install -r services/worker/requirements.txt
python services/worker/app.py
```

In another terminal, start the gateway:

```powershell
go run ./services/gateway
```

Start the frontend:

```powershell
cd frontend
corepack npm install
corepack npm run dev
```

Open the UI at [http://localhost:5173](http://localhost:5173). The gateway defaults to `localhost:8080` and the worker to `localhost:9000`.

## Configuration

Copy `.env.example` to `.env` and set only the services you are using. Keep real credentials out of Git:

```text
DATABASE_URL=postgres://...
TEMPORAL_ADDRESS=localhost:7233
REDIS_URL=redis://localhost:6379/0
KAFKA_BROKERS=localhost:9092
AUTH_JWT_SECRET=...
AGENTMESH_LLM_PROVIDER=mock|openai|anthropic
OPENAI_API_KEY=...
ANTHROPIC_API_KEY=...
```

## Testing everything

Run Go tests:

```powershell
go test ./...
```

Run worker/provider tests:

```powershell
python -m pytest services/worker/tests
```

Run frontend type-check/build tests:

```powershell
cd frontend
corepack npm test
corepack npm run build
```

Validate the Docker Compose model:

```powershell
docker compose -f deploy/docker/docker-compose.yml config
```

Build and run the local containers:

```powershell
docker compose -f deploy/docker/docker-compose.yml up --build
```

Validate Kubernetes/Kustomize when `kubectl` is available:

```powershell
kubectl kustomize deploy/kubernetes
kubectl apply --dry-run=client -k deploy/kubernetes
```

Validate Terraform infrastructure configuration:

```powershell
cd infra/terraform
terraform init
terraform fmt -check
terraform validate
terraform plan
```

Validate and render Helm:

```powershell
helm lint deploy/helm/agentmesh
helm template agentmesh deploy/helm/agentmesh
```

## Phase history

- Phase 0 — repository scaffold and service seams.
- Phase 1 — Go REST gateway to Python worker over gRPC.
- Phase 2 — PostgreSQL run persistence and migrations.
- Phase 3 — Temporal workflows, retries, and asynchronous execution.
- Phase 4 — mock, OpenAI, and Anthropic provider adapters.
- Phase 5 — Vite frontend with run submission, polling, results, and errors.
- Phase 6 — WebSocket run-event streaming with polling fallback.
- Phase 7 — Redis write-through run caching.
- Phase 8 — Kafka run-event publishing.
- Phase 9 — optional JWT authentication.
- Phase 10 — OpenTelemetry instrumentation, logs, and metrics.
- Phase 11 — Dockerfiles, Compose, and Nginx proxying.
- Phase 12 — Kubernetes/Kustomize deployment manifests.
- Phase 13 — Helm chart packaging.
- Phase 14 — AWS infrastructure with Terraform.
- Phase 15 — GitOps with Argo CD and edge routing with Envoy.
- Phase 16 — k6 load testing, failure checks, and production hardening.

All planned phases are complete. Use [docs/SETUP_AND_VERIFICATION.md](docs/SETUP_AND_VERIFICATION.md) to reproduce and verify the entire system on a Docker-capable laptop.

See [docs/SETUP_AND_VERIFICATION.md](docs/SETUP_AND_VERIFICATION.md), [docs/phase-plan.md](docs/phase-plan.md), and the deployment READMEs for details.