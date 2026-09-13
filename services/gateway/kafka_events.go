package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"strings"
	"time"
)

type RunEvent struct {
	EventVersion int       `json:"event_version"`
	EventType    string    `json:"event_type"`
	Run          Run       `json:"run"`
	OccurredAt   time.Time `json:"occurred_at"`
}
type RunEventPublisher interface {
	Publish(context.Context, RunEvent) error
	Close() error
}
type KafkaRunStore struct {
	base      RunStore
	publisher RunEventPublisher
}

func NewKafkaRunStore(base RunStore, brokers, topic string) (*KafkaRunStore, error) {
	if strings.TrimSpace(brokers) == "" || strings.TrimSpace(topic) == "" {
		return nil, fmt.Errorf("kafka brokers and topic are required")
	}
	return &KafkaRunStore{base: base, publisher: &kafkaPublisher{writer: &kafka.Writer{Addr: kafka.TCP(strings.Split(brokers, ",")...), Topic: topic, Balancer: &kafka.LeastBytes{}}}}, nil
}
func (k *KafkaRunStore) Save(run Run) error {
	if err := k.base.Save(run); err != nil {
		return err
	}
	return k.publisher.Publish(context.Background(), RunEvent{EventVersion: 1, EventType: "run.updated", Run: run, OccurredAt: time.Now().UTC()})
}
func (k *KafkaRunStore) Get(id string) (Run, bool, error) { return k.base.Get(id) }
func (k *KafkaRunStore) Close() error                     { return k.publisher.Close() }

type kafkaPublisher struct{ writer *kafka.Writer }

func (p *kafkaPublisher) Publish(ctx context.Context, event RunEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{Key: []byte(event.Run.ID), Value: data})
}
func (p *kafkaPublisher) Close() error { return p.writer.Close() }
