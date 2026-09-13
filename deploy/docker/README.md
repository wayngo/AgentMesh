# Docker local stack

Phase 11 provides a local container stack for the AgentMesh gateway, Python worker, and Vite-built frontend.

```powershell
docker compose -f deploy/docker/docker-compose.yml up --build
```

Open the UI at http://localhost:3000. The gateway is available at http://localhost:8080. The default worker provider is `mock`; infrastructure services from earlier phases can be enabled by setting `DATABASE_URL`, `REDIS_URL`, `TEMPORAL_ADDRESS`, or `KAFKA_BROKERS` in the environment before starting Compose.