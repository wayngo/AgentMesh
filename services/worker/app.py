from concurrent import futures
import grpc
from fastapi import FastAPI
from pydantic import BaseModel
try:
    from .providers import build_provider
except ImportError:
    from providers import build_provider

app = FastAPI(title="AgentMesh Worker")
_provider = build_provider()

class ExecuteRequest(BaseModel):
    run_id: str
    agent_id: str
    input: str = ""

@app.get("/healthz")
def healthz():
    return {"status": "ok"}

@app.post("/internal/execute")
def execute(r: ExecuteRequest):
    return {"run_id": r.run_id, "status": "completed", "result": _provider.generate(r.input)}

def _put_string(out: bytearray, field: int, value: str) -> None:
    if not value: return
    out.append((field << 3) | 2)
    data = value.encode()
    n = len(data)
    while n >= 128:
        out.append((n & 0x7f) | 0x80); n >>= 7
    out.append(n); out.extend(data)

def _marshal(values: tuple[str, str, str]) -> bytes:
    out = bytearray()
    for field, value in enumerate(values, 1): _put_string(out, field, value)
    return bytes(out)

def _unmarshal(data: bytes) -> tuple[str, str, str]:
    values = [""] * 3; i = 0
    while i < len(data):
        tag = data[i]; i += 1
        if tag & 7 != 2: raise ValueError("unsupported protobuf wire type")
        field = tag >> 3; length = 0; shift = 0
        while True:
            if i >= len(data) or shift > 28: raise ValueError("invalid protobuf length")
            c = data[i]; i += 1; length |= (c & 127) << shift
            if c < 128: break
            shift += 7
        end = i + length
        if end > len(data): raise ValueError("invalid protobuf payload")
        if 1 <= field <= 3: values[field - 1] = data[i:end].decode()
        i = end
    return tuple(values)

def execute_run(request_bytes: bytes, context: grpc.ServicerContext) -> bytes:
    run_id, agent_id, input_text = _unmarshal(request_bytes)
    if not agent_id:
        context.abort(grpc.StatusCode.INVALID_ARGUMENT, "agent_id is required")
    try:
        result = _provider.generate(input_text)
    except Exception as exc:
        context.abort(grpc.StatusCode.INTERNAL, f"LLM provider failed: {exc}")
    return _marshal((run_id, "completed", result))

def create_grpc_server() -> grpc.Server:
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    handler = grpc.unary_unary_rpc_method_handler(
        execute_run, request_deserializer=lambda b: b,
        response_serializer=lambda b: b,
    )
    server.add_generic_rpc_handlers((grpc.method_handlers_generic_handler(
        "agentmesh.run.v1.RunWorker", {"ExecuteRun": handler}),))
    return server

def serve_grpc(address: str = "[::]:9000") -> grpc.Server:
    server = create_grpc_server(); server.add_insecure_port(address); server.start(); return server

if __name__ == "__main__":
    server = serve_grpc()
    print("AgentMesh worker gRPC listening on :9000", flush=True)
    server.wait_for_termination()