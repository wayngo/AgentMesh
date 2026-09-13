import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: Number(__ENV.VUS || 5),
  duration: __ENV.DURATION || "30s",
  thresholds: { http_req_failed: ["rate<0.05"], http_req_duration: ["p(95)<1500"] },
};
const base = __ENV.BASE_URL || "http://localhost:8080";

export default function () {
  const create = http.post(`${base}/v1/runs`, JSON.stringify({ agent_id: "load-test", input: `hello-${__VU}-${__ITER}` }), { headers: { "Content-Type": "application/json" } });
  check(create, { "create is successful or queued": (r) => r.status === 200 || r.status === 202 });
  if (create.status === 200 || create.status === 202) {
    const run = create.json();
    check(run, { "run has id": (body) => Boolean(body.id), "run has status": (body) => Boolean(body.status) });
    const get = http.get(`${base}/v1/runs/${encodeURIComponent(run.id)}`);
    check(get, { "get run is successful": (r) => r.status === 200 });
  }
  sleep(0.2);
}