# AgentMesh

AgentMesh is a distributed AI-agent execution platform designed to run reliable, observable, and scalable agent workloads. It separates API ingress, workflow orchestration, model execution, persistence, event distribution, and edge delivery so each layer can scale and evolve independently.

## Platform capabilities

| Technology | Role in AgentMesh |
|---|---|
| Go | High-performance gateway and API lifecycle management. |
| Python / FastAPI | Worker runtime for model execution and provider integration. |
| gRPC / Protocol Buffers | Strongly typed, efficient gateway-to-worker communication through a versioned service contract. |
| Temporal | Durable workflow orchestration, retries, asynchronous execution, and recovery from process or infrastructure failures. |
| PostgreSQL | Durable system-of-record for agent runs, statuses, results, and lifecycle state. |
| Redis | Low-latency write-through caching for frequently accessed run state. |
| Apache Kafka | Versioned event distribution for run-state changes and downstream consumers. |
| OpenTelemetry | Distributed HTTP instrumentation and trace/span propagation for observability. |
| Structured logging / metrics | Machine-readable request logs plus request, error, and latency metrics at `/debug/metrics`. |
| WebSockets | Real-time run-status updates to connected clients. |
| Vite / TypeScript | Typed frontend build pipeline and browser client for submitting and monitoring runs. |
| Docker / Docker Compose | Reproducible local builds and multi-service development environments. |
| Kubernetes | Container scheduling, service discovery, health probes, rolling deployments, and horizontal scaling. |
| Helm | Parameterized Kubernetes packaging for repeatable environment deployments. |
| AWS | Cloud infrastructure target for production workloads. |
| Terraform | Infrastructure as code for repeatable AWS provisioning and reviewable plans. |
| Amazon EKS | Managed Kubernetes control plane and workload platform. |
| Amazon RDS PostgreSQL | Managed relational persistence for production run state. |
| Amazon ElastiCache Redis | Managed cache layer for low-latency state access. |
| Amazon MSK Kafka | Managed event streaming for run lifecycle events. |
| GitOps / Argo CD | Automated Kubernetes synchronization, drift correction, pruning, and self-healing from Git. |
| Envoy | Edge routing and WebSocket-aware traffic handling in front of the application. |
| k6 | Load testing for run throughput, latency, and error-rate regression detection. |

## Architecture

```text
Client / Browser
      |
      v
Envoy or Ingress
      |
      v
Frontend (Vite / TypeScript / Nginx)
      |
      +--> Go Gateway (REST, JWT, WebSockets, OpenTelemetry)
                  |
                  +--> Temporal workflows and activities
                  +--> PostgreSQL / Redis / Kafka
                  +--> Python Worker over gRPC
                              |
                              +--> Mock, OpenAI, or Anthropic provider
```

The gateway accepts run requests and stores lifecycle state. When Temporal is configured, it starts a durable workflow and returns a queued response; the workflow invokes the worker activity with retry policy and persists the final result. Without Temporal, the gateway can execute directly through gRPC for lightweight local development.

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

Open the UI at [http://localhost:5173](http://localhost:5173). The gateway defaults to `http://localhost:8080` and the worker gRPC service to `localhost:9000`.

Direct API checks:

```powershell
Invoke-RestMethod http://localhost:8080/healthz
$body = '{"agent_id":"demo","input":"hello"}'
Invoke-RestMethod -Method Post http://localhost:8080/v1/runs -ContentType 'application/json' -Body $body
Invoke-RestMethod http://localhost:8080/debug/metrics
```

## Configuration

Copy `.env.example` to `.env` and set only the integrations you are using:

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

Use `AGENTMESH_LLM_PROVIDER=mock` for local testing. OpenAI and Anthropic providers require their corresponding API keys and model settings.

## Security

Never commit `.env` files, API keys, database credentials, JWT secrets, Terraform state, AWS/cloud credentials, kubeconfig files, or private certificates. Commit `.env.example` only. Always inspect `git status` and `git diff --cached` before pushing. If a secret is exposed, revoke or rotate it immediately; deleting it later does not remove it from Git history.

## Testing and verification

Run the Go test suite:

```powershell
go test ./...
```

Run worker and provider tests:

```powershell
python -m pytest services/worker/tests
```

Run frontend checks:

```powershell
cd frontend
corepack npm test
corepack npm run build
cd ..
```

Run API failure checks:

```powershell
.\tests\failure\verify-failures.ps1 -BaseUrl http://localhost:8080
```

Run load tests with k6:

```powershell
k6 run tests/k6/runs.js
$env:BASE_URL="http://localhost:8080"
$env:VUS="10"
$env:DURATION="2m"
k6 run tests/k6/runs.js
```

See [docs/SETUP_AND_VERIFICATION.md](docs/SETUP_AND_VERIFICATION.md) for the complete local, Docker, Kubernetes, Helm, AWS, Argo CD, Envoy, failure-testing, and load-testing procedures.

## Container deployment

Validate and run the local Docker stack:

```powershell
docker compose -f deploy/docker/docker-compose.yml config
docker compose -f deploy/docker/docker-compose.yml up --build
```

The frontend is available at `http://localhost:3000` and the gateway at `http://localhost:8080`.

## Kubernetes and GitOps

Render and validate the Kubernetes manifests:

```powershell
kubectl kustomize deploy/kubernetes
kubectl apply --dry-run=client -k deploy/kubernetes
```

Validate and render the Helm chart:

```powershell
helm lint deploy/helm/agentmesh
helm template agentmesh deploy/helm/agentmesh
```

Apply the Argo CD application after replacing the placeholder repository URL in `deploy/argocd/agentmesh-application.yaml`:

```powershell
kubectl apply -f deploy/argocd/agentmesh-application.yaml
```

## AWS infrastructure

Terraform configuration is in `infra/terraform`. Review the plan before creating billable resources:

```powershell
cd infra/terraform
Copy-Item terraform.tfvars.example terraform.tfvars
terraform init
terraform fmt -check
terraform validate
terraform plan
```

Use an encrypted, access-controlled remote Terraform backend before shared or production use. See [infra/terraform/README.md](infra/terraform/README.md) for the full workflow.

## Repository layout

```text
services/gateway/            Go REST gateway and infrastructure integrations
services/worker/             Python worker and LLM provider adapters
frontend/                    TypeScript/Vite browser client
proto/                       Versioned Protocol Buffers contract
db/migrations/               PostgreSQL schema migrations
deploy/docker/               Dockerfiles, Compose, and Nginx proxy
deploy/kubernetes/           Kubernetes/Kustomize resources
deploy/helm/                 Helm chart
deploy/argocd/               Argo CD GitOps resources
deploy/envoy/                Envoy edge configuration
infra/terraform/             AWS infrastructure as code
observability/               Observability documentation and configuration
tests/                       Load and failure testing
docs/                        Setup, verification, and architecture documentation
```

## Architecture diagram

See [docs/architecture-overview.md](docs/architecture-overview.md) for a simple visual request-flow diagram and an explanation of how the platform components work together.

## Documentation

- [Setup and verification guide](docs/SETUP_AND_VERIFICATION.md)
- [Docker deployment](deploy/docker/README.md)
- [Kubernetes deployment](deploy/kubernetes/README.md)
- [Helm deployment](deploy/helm/README.md)
- [Terraform AWS infrastructure](infra/terraform/README.md)
- [Argo CD GitOps](deploy/argocd/README.md)
- [Envoy edge routing](deploy/envoy/README.md)