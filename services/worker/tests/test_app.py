import pytest
from services.worker.app import _marshal, _unmarshal, create_grpc_server

def test_proto_wire_round_trip():
    assert _unmarshal(_marshal(("run-1", "demo", "hello"))) == ("run-1", "demo", "hello")

def test_proto_wire_rejects_truncated_payload():
    with pytest.raises(ValueError): _unmarshal(bytes([0x0A, 0x05, 0x68]))

def test_grpc_server_starts():
    server = create_grpc_server()
    assert server.add_insecure_port("localhost:0") > 0
    server.stop(0)