# GitOps deployment

Phase 15 adds an Argo CD `AppProject` and `Application` that continuously deploy the AgentMesh Helm chart from the configured Git repository. Replace the placeholder repository URL before applying:

```powershell
kubectl apply -f deploy/argocd/agentmesh-application.yaml
```

The application enables automated sync, pruning, and self-healing. Install Argo CD and grant it access to the repository before applying this file.