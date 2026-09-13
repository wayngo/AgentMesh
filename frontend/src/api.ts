export type RunStatus = "queued" | "completed" | "failed" | string;
export type Run = { id: string; agent_id: string; status: RunStatus; result?: string };
export type CreateRunRequest = { agent_id: string; input: string };

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, { headers: { "content-type": "application/json" }, ...init });
  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || `Request failed with ${response.status}`);
  }
  return response.json() as Promise<T>;
}

export function createRun(input: CreateRunRequest): Promise<Run> {
  return request<Run>("/v1/runs", { method: "POST", body: JSON.stringify(input) });
}
export function getRun(id: string): Promise<Run> {
  return request<Run>(`/v1/runs/${encodeURIComponent(id)}`);
}
export async function waitForRun(id: string, options: { intervalMs?: number; timeoutMs?: number } = {}): Promise<Run> {
  const intervalMs = options.intervalMs ?? 1000;
  const deadline = Date.now() + (options.timeoutMs ?? 120000);
  while (true) {
    const run = await getRun(id);
    if (run.status === "completed" || run.status === "failed") return run;
    if (Date.now() >= deadline) throw new Error("Timed out waiting for run");
    await new Promise((resolve) => setTimeout(resolve, intervalMs));
  }
}
export function watchRun(id: string, options: { timeoutMs?: number } = {}): Promise<Run> {
  if (typeof WebSocket === "undefined") return waitForRun(id, options);
  return new Promise((resolve, reject) => {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    const socket = new WebSocket(`${protocol}//${window.location.host}/v1/runs/${encodeURIComponent(id)}/events`);
    const timeout = window.setTimeout(() => { socket.close(); reject(new Error("Timed out waiting for run")); }, options.timeoutMs ?? 120000);
    socket.onmessage = (event) => {
      const run = JSON.parse(event.data) as Run;
      if (run.status === "completed" || run.status === "failed") { window.clearTimeout(timeout); socket.close(); resolve(run); }
    };
    socket.onerror = () => { window.clearTimeout(timeout); socket.close(); waitForRun(id, options).then(resolve, reject); };
  });
}