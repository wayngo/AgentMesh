from types import SimpleNamespace
import pytest
from services.worker.providers import AnthropicProvider, MockProvider, OpenAIProvider, build_provider

def test_mock_provider_is_default(monkeypatch):
    monkeypatch.delenv("AGENTMESH_LLM_PROVIDER", raising=False)
    assert build_provider().generate("hello") == "mock response: hello"

def test_openai_adapter_uses_responses_api():
    calls=[]
    client=SimpleNamespace(responses=SimpleNamespace(create=lambda **kwargs: (calls.append(kwargs) or SimpleNamespace(output_text="answer"))))
    provider=OpenAIProvider(client,"test-model")
    assert provider.generate("hello")=="answer"
    assert calls==[{"model":"test-model","input":"hello"}]

def test_anthropic_adapter_uses_messages_api():
    calls=[]
    response=SimpleNamespace(content=[SimpleNamespace(text="answer")])
    client=SimpleNamespace(messages=SimpleNamespace(create=lambda **kwargs: (calls.append(kwargs) or response)))
    provider=AnthropicProvider(client,"test-model")
    assert provider.generate("hello")=="answer"
    assert calls==[{"model":"test-model","max_tokens":1024,"messages":[{"role":"user","content":"hello"}]}]

def test_unknown_provider_is_rejected():
    with pytest.raises(ValueError, match="unsupported LLM provider"):
        build_provider("unknown")