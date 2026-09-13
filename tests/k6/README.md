# k6 load tests

Phase 16 load test for the gateway run lifecycle.

```powershell
k6 run tests/k6/runs.js
$env:BASE_URL="http://localhost:8080"; $env:VUS="10"; $env:DURATION="2m"; k6 run tests/k6/runs.js
```

The script accepts both synchronous `200` and Temporal-backed asynchronous `202` run creation. It checks run IDs, statuses, retrieval, error rate, and p95 latency.