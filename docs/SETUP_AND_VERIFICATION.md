# AgentMesh Setup and Verification Guide

This guide is the reproducible handoff for all completed phases. It is designed for running the project on a laptop that has Docker, Kubernetes, Helm, Terraform, and k6 available. The current development device does not need to run Docker; copy the repository to the laptop and run the container, Kubernetes, and cloud checks there.

## 1. What has been built

- Phase 0 — scaffold: repository layout, Go gateway, Python worker, protobuf contract, frontend shell, CI/deployment directories.
- Phase 1 — gRPC execution: REST gateway calls the Python worker through the `RunWorker/ExecuteRun` contract.
- Phase 2 — PostgreSQL: durable `runs` storage, migration `db/migrations/001_runs.sql`, and memory fallback.
- Phase 3 — Temporal: asynchronous durable workflows with retrying activities when `TEMPORAL_ADDRESS` is configured.
- Phase 4 — LLM providers: mock, OpenAI, and Anthropic adapters selected by `AGENTMESH_LLM_PROVIDER`.
- Phase 5 — frontend: Vite UI for submitting runs, waiting for completion, and displaying results/errors.
- Phase 6 — WebSockets: `/v1/runs/{id}/events` live updates with frontend polling fallback.
- Phase 7 — Redis: optional write-through cache with source-store fallback.
- Phase 8 — Kafka: optional versioned `run.updated` events.
- Phase 9 — authentication: optional HS256 JWT bearer authentication.
- Phase 10 — observability: OpenTelemetry HTTP spans, structured logs, and `/debug/metrics`.
- Phase 11 — Docker: gateway, worker, frontend images and local Compose stack.
- Phase 12 — Kubernetes: Kustomize deployments, services, probes, secrets, and ingress.
- Phase 13 — Helm: configurable Helm chart for the Kubernetes stack.
- Phase 14 — AWS: Terraform VPC, EKS, RDS, ElastiCache, and MSK foundation.
- Phase 15 — GitOps/edge: Argo CD application and Envoy WebSocket-aware edge routing.
- Phase 16 — validation: k6 load testing, failure checks, and production-hardening checklist.

## 2. Laptop prerequisites

Install and verify:

```powershell
go version
python --version
corepack npm --version
docker version
docker compose version
kubectl version --client
helm version
terraform version
k6 version
```

The minimum application dependencies are Go 1.23+, Python 3.13+, Node.js 22+, Docker Desktop, kubectl, Helm 3, Terraform 1.6+, and k6. Kubernetes and Terraform checks require their corresponding tools; they cannot be fully verified from a machine without Docker/Kubernetes/AWS access.

## 3. Local non-Docker verification

From the repository root:

```powershell
go test ./...
python -m pip install -r services/worker/requirements.txt
python -m pytest services/worker/tests
cd frontend
corepack npm install
corepack npm test
corepack npm run build
cd ..
```

Expected result: Go tests pass, Python reports 7 passed, and the frontend TypeScript/Vite build succeeds.

Start the worker in terminal 1:

```powershell
python services/worker/app.py
```

Start the gateway in terminal 2:

```powershell
go run ./services/gateway
```

Start the frontend in terminal 3:

```powershell
cd frontend
corepack npm run dev
```

Open `http://localhost:5173`. Submit a run using the default `demo` agent. The gateway API is `http://localhost:8080`; the worker gRPC port is `9000`.

Direct API checks:

```powershell
Invoke-RestMethod http://localhost:8080/healthz
$body = '{"agent_id":"demo","input":"hello"}'
Invoke-RestMethod -Method Post http://localhost:8080/v1/runs -ContentType 'application/json' -Body $body
Invoke-RestMethod http://localhost:8080/debug/metrics
```

## 4. Optional service configuration

Copy `.env.example` to `.env` and set only the integrations that exist on the laptop. Do not commit `.env` or real credentials.

- `DATABASE_URL` enables PostgreSQL persistence.
- `TEMPORAL_ADDRESS` enables asynchronous Temporal workflows.
- `REDIS_URL` enables caching.
- `KAFKA_BROKERS` and `KAFKA_TOPIC` enable Kafka events.
- `AUTH_JWT_SECRET` enables JWT protection for run endpoints.
- `AGENTMESH_LLM_PROVIDER=mock|openai|anthropic` selects the worker provider.
- `OPENAI_API_KEY` and `ANTHROPIC_API_KEY` are required only for their providers.

Use `mock` first. It verifies the full application path without paid API calls.

## 5. Docker verification

From the repository root on the laptop:

```powershell
docker compose -f deploy/docker/docker-compose.yml config
docker compose -f deploy/docker/docker-compose.yml build
docker compose -f deploy/docker/docker-compose.yml up
```

Expected result: `frontend` is available at `http://localhost:3000`, `gateway` at `http://localhost:8080`, and the gateway health check becomes healthy. Open the frontend and submit a mock run.

In another terminal:

```powershell
Invoke-RestMethod http://localhost:8080/healthz
Invoke-RestMethod http://localhost:3000
```

Stop the stack:

```powershell
docker compose -f deploy/docker/docker-compose.yml down
```

If Docker build fails, inspect Docker Desktop login, disk space, WSL2/backend status, and permissions. The gateway image downloads Go modules during its build and the frontend image downloads npm packages.

## 6. Kubernetes/Kustomize verification

A local Kubernetes cluster can be Docker Desktop Kubernetes, minikube, or kind. Build images and load them into the cluster as required by the chosen tool.

```powershell
docker build -f deploy/docker/Dockerfile.gateway -t agentmesh/gateway:local .
docker build -f deploy/docker/Dockerfile.worker -t agentmesh/worker:local .
docker build -f deploy/docker/Dockerfile.frontend -t agentmesh/frontend:local .
kubectl kustomize deploy/kubernetes
kubectl apply --dry-run=client -k deploy/kubernetes
kubectl apply -k deploy/kubernetes
kubectl -n agentmesh get pods,svc,ingress
kubectl -n agentmesh rollout status deployment/gateway
kubectl -n agentmesh rollout status deployment/worker
kubectl -n agentmesh rollout status deployment/frontend
```

Expected result: all pods become Ready; gateway and frontend services have endpoints. Configure `agentmesh.local` to the ingress address or use port forwarding:

```powershell
kubectl -n agentmesh port-forward service/frontend 3000:80
```

Never apply `secret.example.yaml` with real production credentials committed to Git. Use a secret manager or environment-specific overlay.

## 7. Helm verification

```powershell
helm lint deploy/helm/agentmesh
helm template agentmesh deploy/helm/agentmesh > rendered-agentmesh.yaml
kubectl apply --dry-run=client -f rendered-agentmesh.yaml
helm upgrade --install agentmesh deploy/helm/agentmesh --namespace agentmesh --create-namespace
helm status agentmesh --namespace agentmesh
```

Use a private `values.local.yaml` for real image tags, hostnames, replica counts, and secret references. Review rendered YAML before installing.

## 8. Terraform/AWS verification

Terraform provisions billable resources. Never run `apply` until the plan is reviewed and AWS credentials, account, region, and budget are confirmed.

```powershell
cd infra/terraform
Copy-Item terraform.tfvars.example terraform.tfvars
terraform init
terraform fmt -check
terraform validate
terraform plan -out agentmesh.tfplan
terraform show agentmesh.tfplan
```

Only after review:

```powershell
terraform apply agentmesh.tfplan
terraform output
aws eks update-kubeconfig --region us-west-2 --name (terraform output -raw eks_cluster_name)
```

Verify EKS access, then deploy Helm. Store Terraform state in an encrypted, locked remote backend before shared use. The generated RDS password is in Terraform state, so state access must be restricted.

## 9. Argo CD and Envoy verification

Replace the placeholder Git repository URL in `deploy/argocd/agentmesh-application.yaml`, then:

```powershell
kubectl apply -f deploy/argocd/agentmesh-application.yaml
kubectl -n argocd get applications
kubectl -n argocd describe application agentmesh
```

Expected result: Argo CD reports Synced and Healthy. Verify Envoy configuration before deployment and confirm WebSocket upgrades by submitting a run from the frontend while watching the run status.

## 10. Failure and load verification

Run API failure checks:

```powershell
.\tests\failure\verify-failures.ps1 -BaseUrl http://localhost:8080
```

Expected result: missing `agent_id` returns 400 and an unknown run returns 404. With JWT enabled, repeat with `-JwtToken` and verify unauthenticated requests return 401.

Run the load test:

```powershell
k6 run tests/k6/runs.js
$env:BASE_URL="http://localhost:8080"
$env:VUS="10"
$env:DURATION="2m"
k6 run tests/k6/runs.js
```

The default thresholds are fewer than 5% failed requests and p95 latency below 1.5 seconds. Record CPU/memory, gateway errors, worker errors, database connections, Redis hit behavior, Kafka lag, and Temporal retries while the test runs.

## 11. Final readiness checklist

- All Go and Python tests pass.
- Frontend type-check and production build pass.
- Docker Compose config/build/startup verified on the laptop.
- Kubernetes manifests render and deployments become Ready.
- Helm lint/template/install verified.
- Terraform validate/plan reviewed; apply approved separately.
- Secrets are externalized and not committed.
- JWT authentication is enabled outside local mock development.
- TLS is configured at the production edge.
- OpenTelemetry collector/exporter is configured.
- PostgreSQL backups and restore procedure are tested.
- Redis and Kafka failure behavior is understood.
- Temporal retry and duplicate workflow behavior is tested.
- k6 load and failure results are recorded.

For the phase-level summary, see [docs/phase-plan.md](phase-plan.md). Deployment-specific details are in the READMEs under `deploy/` and `infra/terraform/`.