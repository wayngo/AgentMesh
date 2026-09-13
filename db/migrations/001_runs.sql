-- AgentMesh Phase 2: durable run state.
CREATE TABLE IF NOT EXISTS runs (
    id TEXT PRIMARY KEY,
    agent_id TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('queued', 'completed', 'failed')),
    result TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS runs_agent_id_idx ON runs (agent_id);
CREATE INDEX IF NOT EXISTS runs_status_idx ON runs (status);