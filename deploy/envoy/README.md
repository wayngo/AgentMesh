# Envoy edge routing

`envoy.yaml` defines an HTTP edge listener on port 8080 that routes to the frontend Kubernetes service and supports WebSocket upgrades. The frontend continues to proxy `/v1` and `/debug` traffic to the gateway service.

Use this configuration with an Envoy deployment or Gateway API controller in the cluster. TLS termination should be added at the environment edge before production use.