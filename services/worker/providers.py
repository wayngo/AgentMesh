from __future__ import annotations

import os
from dataclasses import dataclass
from typing import Any, Protocol

class LLMProvider(Protocol):
    def generate(self, prompt: str) -> str: ...

@dataclass
class MockProvider:
    prefix: str = "mock response: "
    def generate(self, prompt: str) -> str:
        return f"{self.prefix}{prompt}"

class OpenAIProvider:
    def __init__(self, client: Any, model: str): self.client, self.model = client, model
    def generate(self, prompt: str) -> str:
        response = self.client.responses.create(model=self.model, input=prompt)
        return response.output_text

class AnthropicProvider:
    def __init__(self, client: Any, model: str): self.client, self.model = client, model
    def generate(self, prompt: str) -> str:
        response = self.client.messages.create(model=self.model, max_tokens=1024, messages=[{"role":"user", "content":prompt}])
        return response.content[0].text

def build_provider(provider: str | None = None, *, openai_client: Any = None, anthropic_client: Any = None) -> LLMProvider:
    name = (provider or os.getenv("AGENTMESH_LLM_PROVIDER", "mock")).strip().lower()
    if name == "mock": return MockProvider()
    if name == "openai":
        if openai_client is None:
            from openai import OpenAI
            openai_client = OpenAI(api_key=os.environ.get("OPENAI_API_KEY"))
        return OpenAIProvider(openai_client, os.getenv("OPENAI_MODEL", "gpt-4o-mini"))
    if name == "anthropic":
        if anthropic_client is None:
            from anthropic import Anthropic
            anthropic_client = Anthropic(api_key=os.environ.get("ANTHROPIC_API_KEY"))
        return AnthropicProvider(anthropic_client, os.getenv("ANTHROPIC_MODEL", "claude-3-5-haiku-latest"))
    raise ValueError(f"unsupported LLM provider: {name}")