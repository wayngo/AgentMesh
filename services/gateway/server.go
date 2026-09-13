package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	store    RunStore
	executor RunExecutor
}

func NewServer(st RunStore, ex RunExecutor) *Server { return &Server{st, ex} }
func (s *Server) Handler() http.Handler {
	m := http.NewServeMux()
	m.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	m.HandleFunc("/v1/runs", s.create)
	m.HandleFunc("/v1/runs/", s.runRoute)
	return m
}
func (s *Server) create(w http.ResponseWriter, r *http.Request) {
	var in CreateRunRequest
	if r.Method != http.MethodPost || json.NewDecoder(r.Body).Decode(&in) != nil || strings.TrimSpace(in.AgentID) == "" {
		http.Error(w, "invalid request", 400)
		return
	}
	run := Run{ID: newRunID(), AgentID: in.AgentID, Status: "queued"}
	if err := s.store.Save(run); err != nil {
		http.Error(w, "storage unavailable", 503)
		return
	}
	if async, ok := s.executor.(AsyncRunExecutor); ok {
		if err := async.Start(r.Context(), run, in.Input); err != nil {
			http.Error(w, "workflow unavailable", 503)
			return
		}
		w.WriteHeader(202)
		_ = json.NewEncoder(w).Encode(run)
		return
	}
	out, err := s.executor.Execute(r.Context(), run, in.Input)
	if err != nil {
		run.Status = "failed"
		_ = s.store.Save(run)
		http.Error(w, "execution failed", 502)
		return
	}
	run.Status = "completed"
	run.Result = out
	if err := s.store.Save(run); err != nil {
		http.Error(w, "storage unavailable", 503)
		return
	}
	_ = json.NewEncoder(w).Encode(run)
}
func (s *Server) runRoute(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v1/runs/")
	if strings.HasSuffix(path, "/events") {
		s.events(w, r, strings.TrimSuffix(path, "/events"))
		return
	}
	s.get(w, r, path)
}
func (s *Server) get(w http.ResponseWriter, r *http.Request, id string) {
	run, ok, err := s.store.Get(id)
	if err != nil {
		http.Error(w, "storage unavailable", 503)
		return
	}
	if !ok {
		http.Error(w, "not found", 404)
		return
	}
	_ = json.NewEncoder(w).Encode(run)
}

var wsUpgrader = websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

func (s *Server) events(w http.ResponseWriter, r *http.Request, id string) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	var last Run
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		run, ok, err := s.store.Get(id)
		if err != nil {
			_ = conn.WriteJSON(map[string]string{"error": "storage unavailable"})
			return
		}
		if !ok {
			_ = conn.WriteJSON(map[string]string{"error": "not found"})
			return
		}
		if run != last {
			if err := conn.WriteJSON(run); err != nil {
				return
			}
			last = run
		}
		if run.Status == "completed" || run.Status == "failed" {
			return
		}
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
		}
	}
}
