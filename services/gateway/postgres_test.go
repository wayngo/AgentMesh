package main

import (
	"github.com/DATA-DOG/go-sqlmock"
	"regexp"
	"testing"
)

func TestPostgresRunStoreSaveAndGet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	store := &PostgresRunStore{db: db}
	run := Run{ID: "run-1", AgentID: "demo", Status: "completed", Result: "ok"}
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO runs")).WithArgs(run.ID, run.AgentID, run.Status, run.Result).WillReturnResult(sqlmock.NewResult(1, 1))
	if err := store.Save(run); err != nil {
		t.Fatal(err)
	}
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, agent_id, status, COALESCE(result, '') FROM runs WHERE id=$1")).WithArgs("run-1").WillReturnRows(sqlmock.NewRows([]string{"id", "agent_id", "status", "coalesce"}).AddRow("run-1", "demo", "completed", "ok"))
	got, ok, err := store.Get("run-1")
	if err != nil || !ok || got != run {
		t.Fatalf("got=%+v ok=%v err=%v", got, ok, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPostgresMigrationContainsRunsTable(t *testing.T) {
	if len(postgresSchema) == 0 || !regexp.MustCompile(`CREATE TABLE IF NOT EXISTS runs`).MatchString(postgresSchema) {
		t.Fatal("runs schema missing")
	}
}
