# Helm chart

Phase 13 packages the AgentMesh Kubernetes deployment as `deploy/helm/agentmesh`.

```powershell
helm lint deploy/helm/agentmesh
helm template agentmesh deploy/helm/agentmesh
helm upgrade --install agentmesh deploy/helm/agentmesh --namespace agentmesh --create-namespace
```

Set environment-specific images and credentials through a values overlay; do not commit real secrets. Example:

```powershell
helm upgrade --install agentmesh deploy/helm/agentmesh -f values.local.yaml --namespace agentmesh --create-namespace
```