package main

import (
	"github.com/alicebob/miniredis/v2"
	"testing"
)

func TestRedisRunStoreCachesAndUpdatesRuns(t *testing.T) {
	mini := miniredis.RunT(t)
	base := NewMemoryRunStore()
	store, err := NewRedisRunStore(base, mini.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	run := Run{ID: "redis-1", AgentID: "demo", Status: "queued"}
	if err := store.Save(run); err != nil {
		t.Fatal(err)
	}
	_ = base.Save(Run{ID: run.ID, AgentID: run.AgentID, Status: "completed", Result: "stale-base"})
	got, ok, err := store.Get(run.ID)
	if err != nil || !ok || got.Status != "queued" {
		t.Fatalf("cached got=%+v ok=%v err=%v", got, ok, err)
	}
	run.Status = "completed"
	run.Result = "done"
	if err := store.Save(run); err != nil {
		t.Fatal(err)
	}
	got, ok, err = store.Get(run.ID)
	if err != nil || !ok || got.Result != "done" {
		t.Fatalf("updated got=%+v ok=%v err=%v", got, ok, err)
	}
}

func TestRedisRunStoreFallsBackToBaseOnCacheMiss(t *testing.T) {
	mini := miniredis.RunT(t)
	base := NewMemoryRunStore()
	store, err := NewRedisRunStore(base, mini.Addr())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	run := Run{ID: "redis-2", AgentID: "demo", Status: "completed", Result: "from-base"}
	_ = base.Save(run)
	got, ok, err := store.Get(run.ID)
	if err != nil || !ok || got != run {
		t.Fatalf("got=%+v ok=%v err=%v", got, ok, err)
	}
}
