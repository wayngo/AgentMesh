package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"sync"
	"sync/atomic"
)

type CreateRunRequest struct {
	AgentID string `json:"agent_id"`
	Input   string `json:"input"`
}
type Run struct {
	ID      string `json:"id"`
	AgentID string `json:"agent_id"`
	Status  string `json:"status"`
	Result  string `json:"result,omitempty"`
}
type RunStore interface {
	Save(Run) error
	Get(string) (Run, bool, error)
}
type RunExecutor interface {
	Execute(context.Context, Run, string) (string, error)
}
type MemoryRunStore struct {
	mu   sync.RWMutex
	runs map[string]Run
}

func NewMemoryRunStore() *MemoryRunStore { return &MemoryRunStore{runs: map[string]Run{}} }
func (m *MemoryRunStore) Save(r Run) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runs[r.ID] = r
	return nil
}
func (m *MemoryRunStore) Get(id string) (Run, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	r, ok := m.runs[id]
	return r, ok, nil
}

type MockExecutor struct{}

func NewMockExecutor() MockExecutor { return MockExecutor{} }
func (MockExecutor) Execute(_ context.Context, _ Run, input string) (string, error) {
	return fmt.Sprintf("mock response: %s", input), nil
}

var seq uint64

func newRunID() string { return fmt.Sprintf("run-%d", atomic.AddUint64(&seq, 1)) }

// PostgresRunStore persists run lifecycle state in PostgreSQL.
type PostgresRunStore struct{ db *sql.DB }

func NewPostgresRunStore(ctx context.Context, databaseURL string) (*PostgresRunStore, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	store := &PostgresRunStore{db: db}
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	if _, err := db.ExecContext(ctx, postgresSchema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply postgres schema: %w", err)
	}
	return store, nil
}
func (p *PostgresRunStore) Close() error { return p.db.Close() }
func (p *PostgresRunStore) Save(r Run) error {
	_, err := p.db.Exec(`INSERT INTO runs (id, agent_id, status, result) VALUES ($1,$2,$3,$4) ON CONFLICT (id) DO UPDATE SET agent_id=EXCLUDED.agent_id,status=EXCLUDED.status,result=EXCLUDED.result`, r.ID, r.AgentID, r.Status, r.Result)
	return err
}
func (p *PostgresRunStore) Get(id string) (Run, bool, error) {
	var r Run
	err := p.db.QueryRow(`SELECT id, agent_id, status, COALESCE(result, '') FROM runs WHERE id=$1`, id).Scan(&r.ID, &r.AgentID, &r.Status, &r.Result)
	if err == sql.ErrNoRows {
		return Run{}, false, nil
	}
	if err != nil {
		return Run{}, false, err
	}
	return r, true, nil
}
