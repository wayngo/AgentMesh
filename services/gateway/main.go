package main

import (
	"context"
	"expvar"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	addr := os.Getenv("AGENTMESH_GATEWAY_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	workerAddr := os.Getenv("AGENTMESH_WORKER_ADDR")
	if workerAddr == "" {
		workerAddr = "localhost:9000"
	}
	grpcExecutor, err := NewGRPCExecutor(workerAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer grpcExecutor.Close()
	var store RunStore = NewMemoryRunStore()
	var closeStore func()
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		postgres, err := NewPostgresRunStore(context.Background(), databaseURL)
		if err != nil {
			log.Fatal(err)
		}
		store = postgres
		closeStore = func() { _ = postgres.Close() }
	}
	if closeStore != nil {
		defer closeStore()
	}
	if redisAddress := os.Getenv("REDIS_URL"); redisAddress != "" {
		cached, err := NewRedisRunStore(store, redisAddress)
		if err != nil {
			log.Fatal(err)
		}
		store = cached
		defer cached.Close()
	}
	if brokers := os.Getenv("KAFKA_BROKERS"); strings.TrimSpace(brokers) != "" {
		topic := os.Getenv("KAFKA_TOPIC")
		if topic == "" {
			topic = "agentmesh.run-events"
		}
		events, err := NewKafkaRunStore(store, brokers, topic)
		if err != nil {
			log.Fatal(err)
		}
		store = events
		defer events.Close()
		log.Printf("kafka events enabled topic=%s", topic)
	}
	var executor RunExecutor = grpcExecutor
	if temporalAddress := os.Getenv("TEMPORAL_ADDRESS"); temporalAddress != "" {
		queue := os.Getenv("TEMPORAL_TASK_QUEUE")
		if queue == "" {
			queue = "agentmesh-runs"
		}
		runtime, err := NewTemporalRuntime(context.Background(), temporalAddress, queue, &RunActivities{Store: store, Executor: grpcExecutor})
		if err != nil {
			log.Fatal(err)
		}
		defer runtime.Close()
		executor = NewTemporalRunExecutor(runtime.client, queue)
	}
	apiHandler := ObservabilityMiddleware(NewServer(store, executor).Handler())
	var handler http.Handler = http.NewServeMux()
	handler.(*http.ServeMux).Handle("/debug/metrics", expvar.Handler())
	handler.(*http.ServeMux).Handle("/", apiHandler)
	if secret := os.Getenv("AUTH_JWT_SECRET"); secret != "" {
		handler = NewAuthMiddleware(secret).Wrap(handler)
		log.Printf("jwt authentication enabled")
	}
	log.Printf("gateway listening on %s; worker=%s; persistence=%T", addr, workerAddr, store)
	log.Fatal(http.ListenAndServe(addr, handler))
}
