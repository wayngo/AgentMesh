package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"net/url"
	"strings"
	"time"
)

type RedisRunStore struct {
	base   RunStore
	client *redis.Client
	ttl    time.Duration
}

func NewRedisRunStore(base RunStore, address string) (*RedisRunStore, error) {
	options := &redis.Options{Addr: address}
	if strings.Contains(address, "://") {
		parsed, err := url.Parse(address)
		if err != nil {
			return nil, err
		}
		options = &redis.Options{Addr: parsed.Host, Password: "", DB: 0}
		if parsed.User != nil {
			options.Password, _ = parsed.User.Password()
		}
	}
	client := redis.NewClient(options)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &RedisRunStore{base: base, client: client, ttl: 15 * time.Minute}, nil
}
func (r *RedisRunStore) key(id string) string { return "agentmesh:run:" + id }
func (r *RedisRunStore) Save(run Run) error {
	if err := r.base.Save(run); err != nil {
		return err
	}
	data, err := json.Marshal(run)
	if err != nil {
		return err
	}
	return r.client.Set(context.Background(), r.key(run.ID), data, r.ttl).Err()
}
func (r *RedisRunStore) Get(id string) (Run, bool, error) {
	data, err := r.client.Get(context.Background(), r.key(id)).Bytes()
	if err == nil {
		var run Run
		if json.Unmarshal(data, &run) == nil {
			return run, true, nil
		}
	} else if err != redis.Nil {
		return Run{}, false, err
	}
	run, ok, err := r.base.Get(id)
	if err != nil || !ok {
		return run, ok, err
	}
	if data, marshalErr := json.Marshal(run); marshalErr == nil {
		_ = r.client.Set(context.Background(), r.key(id), data, r.ttl).Err()
	}
	return run, true, nil
}
func (r *RedisRunStore) Close() error { return r.client.Close() }
