# Kubernetes manifests

Phase 12 provides a Kustomize-based deployment for the AgentMesh frontend, gateway, and worker.

Build the application images first:

```powershell
docker build -f deploy/docker/Dockerfile.gateway -t agentmesh/gateway:local .
docker build -f deploy/docker/Dockerfile.worker -t agentmesh/worker:local .
docker build -f deploy/docker/Dockerfile.frontend -t agentmesh/frontend:local .
```

Apply the manifests:

```powershell
kubectl apply -k deploy/kubernetes
```

For production, replace `secret.example.yaml` through an environment overlay or external secret manager. The default host is `agentmesh.local`; configure DNS or `/etc/hosts` for the ingress controller. The ingress keeps connections open for WebSocket run events.