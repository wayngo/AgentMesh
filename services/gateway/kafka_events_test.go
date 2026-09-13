package main

import (
	"context"
	"testing"
)

type fakeRunPublisher struct{ events []RunEvent }

func (p *fakeRunPublisher) Publish(_ context.Context, event RunEvent) error {
	p.events = append(p.events, event)
	return nil
}
func (p *fakeRunPublisher) Close() error { return nil }

func TestKafkaRunStorePublishesRunUpdates(t *testing.T) {
	publisher := &fakeRunPublisher{}
	store := &KafkaRunStore{base: NewMemoryRunStore(), publisher: publisher}
	run := Run{ID: "kafka-1", AgentID: "demo", Status: "completed", Result: "ok"}
	if err := store.Save(run); err != nil {
		t.Fatal(err)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("events=%d", len(publisher.events))
	}
	event := publisher.events[0]
	if event.EventVersion != 1 || event.EventType != "run.updated" || event.Run != run {
		t.Fatalf("event=%+v", event)
	}
}

func TestKafkaRunStoreDoesNotPublishWhenBaseSaveFails(t *testing.T) {
	publisher := &fakeRunPublisher{}
	store := &KafkaRunStore{base: failingRunStore{}, publisher: publisher}
	if err := store.Save(Run{ID: "bad"}); err == nil {
		t.Fatal("expected base error")
	}
	if len(publisher.events) != 0 {
		t.Fatal("published event for failed save")
	}
}

type failingRunStore struct{}

func (failingRunStore) Save(Run) error                { return context.Canceled }
func (failingRunStore) Get(string) (Run, bool, error) { return Run{}, false, context.Canceled }
