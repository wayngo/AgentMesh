package main

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type recordingExecutor struct{ err error }

func (e recordingExecutor) Execute(context.Context, Run, string) (string, error) {
	if e.err != nil {
		return "", e.err
	}
	return "worker result", nil
}

func TestCreateRunPersistsCompletedWorkerResult(t *testing.T) {
	store := NewMemoryRunStore()
	req := httptest.NewRequest(http.MethodPost, "/v1/runs", strings.NewReader(`{"agent_id":"demo","input":"hello"}`))
	res := httptest.NewRecorder()
	NewServer(store, recordingExecutor{}).Handler().ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.Code)
	}
	if !strings.Contains(res.Body.String(), `"result":"worker result"`) {
		t.Fatalf("body = %s", res.Body.String())
	}
}

func TestCreateRunMarksExecutionFailure(t *testing.T) {
	store := NewMemoryRunStore()
	req := httptest.NewRequest(http.MethodPost, "/v1/runs", strings.NewReader(`{"agent_id":"demo","input":"hello"}`))
	res := httptest.NewRecorder()
	NewServer(store, recordingExecutor{err: errors.New("worker unavailable")}).Handler().ServeHTTP(res, req)
	if res.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", res.Code)
	}
	var found bool
	for _, id := range []string{"run-1", "run-2", "run-3"} {
		if run, ok, err := store.Get(id); err == nil && ok && run.Status == "failed" {
			found = true
		}
	}
	if !found {
		t.Fatal("failed run was not persisted")
	}
}

func TestProtoWireRoundTrip(t *testing.T) {
	codec := protoWireCodec{}
	in := &ExecuteRunRequest{RunID: "run-1", AgentID: "demo", Input: "hello"}
	data, err := codec.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out ExecuteRunRequest
	if err := codec.Unmarshal(data, &out); err != nil {
		t.Fatal(err)
	}
	if out != *in {
		t.Fatalf("got %+v, want %+v", out, *in)
	}
}

type asyncExecutor struct{ started bool }

func (e *asyncExecutor) Start(context.Context, Run, string) error             { e.started = true; return nil }
func (e *asyncExecutor) Execute(context.Context, Run, string) (string, error) { return "", nil }

func TestCreateRunStartsTemporalWorkflowAsynchronously(t *testing.T) {
	executor := &asyncExecutor{}
	req := httptest.NewRequest(http.MethodPost, "/v1/runs", strings.NewReader(`{"agent_id":"demo","input":"hello"}`))
	res := httptest.NewRecorder()
	NewServer(NewMemoryRunStore(), executor).Handler().ServeHTTP(res, req)
	if res.Code != http.StatusAccepted || !executor.started {
		t.Fatalf("status=%d started=%v", res.Code, executor.started)
	}
}

func TestRunEventsStreamsUntilCompletion(t *testing.T) {
	store := NewMemoryRunStore()
	run := Run{ID: "run-events", AgentID: "demo", Status: "queued"}
	_ = store.Save(run)
	httpServer := httptest.NewServer(NewServer(store, recordingExecutor{}).Handler())
	defer httpServer.Close()
	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http") + "/v1/runs/run-events/events"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	var first Run
	if err := conn.ReadJSON(&first); err != nil {
		t.Fatal(err)
	}
	if first.Status != "queued" {
		t.Fatalf("first status=%s", first.Status)
	}
	go func() {
		time.Sleep(30 * time.Millisecond)
		run.Status = "completed"
		run.Result = "done"
		_ = store.Save(run)
	}()
	var final Run
	if err := conn.ReadJSON(&final); err != nil {
		t.Fatal(err)
	}
	if final.Status != "completed" || final.Result != "done" {
		t.Fatalf("final=%+v", final)
	}
}
