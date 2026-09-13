package main

import (
	"context"
	"fmt"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
	"time"
)

type RunWorkflowInput struct {
	Run   Run
	Input string
}
type RunWorkflowResult struct {
	Run    Run
	Result string
}
type AsyncRunExecutor interface {
	Start(context.Context, Run, string) error
}
type TemporalRunExecutor struct {
	client    client.Client
	taskQueue string
}

func NewTemporalRunExecutor(c client.Client, taskQueue string) *TemporalRunExecutor {
	return &TemporalRunExecutor{client: c, taskQueue: taskQueue}
}
func (e *TemporalRunExecutor) Start(ctx context.Context, run Run, input string) error {
	_, err := e.client.ExecuteWorkflow(ctx, client.StartWorkflowOptions{ID: run.ID, TaskQueue: e.taskQueue}, RunWorkflow, RunWorkflowInput{Run: run, Input: input})
	return err
}
func (e *TemporalRunExecutor) Execute(context.Context, Run, string) (string, error) {
	return "", fmt.Errorf("temporal executor must be started asynchronously")
}
func RunWorkflow(ctx workflow.Context, input RunWorkflowInput) (RunWorkflowResult, error) {
	opts := workflow.ActivityOptions{StartToCloseTimeout: 2 * time.Minute, RetryPolicy: &temporal.RetryPolicy{InitialInterval: time.Second, BackoffCoefficient: 2, MaximumInterval: 30 * time.Second, MaximumAttempts: 3}}
	ctx = workflow.WithActivityOptions(ctx, opts)
	var result RunWorkflowResult
	err := workflow.ExecuteActivity(ctx, "ExecuteRunActivity", input).Get(ctx, &result)
	return result, err
}

type RunActivities struct {
	Store    RunStore
	Executor RunExecutor
}

func (a *RunActivities) ExecuteRunActivity(ctx context.Context, input RunWorkflowInput) (RunWorkflowResult, error) {
	result, err := a.Executor.Execute(ctx, input.Run, input.Input)
	if err != nil {
		input.Run.Status = "failed"
		_ = a.Store.Save(input.Run)
		return RunWorkflowResult{Run: input.Run}, err
	}
	input.Run.Status = "completed"
	input.Run.Result = result
	if err := a.Store.Save(input.Run); err != nil {
		return RunWorkflowResult{Run: input.Run}, err
	}
	return RunWorkflowResult{Run: input.Run, Result: result}, nil
}

type TemporalRuntime struct {
	client client.Client
	worker worker.Worker
}

func NewTemporalRuntime(ctx context.Context, address, taskQueue string, activities *RunActivities) (*TemporalRuntime, error) {
	c, err := client.Dial(client.Options{HostPort: address})
	if err != nil {
		return nil, fmt.Errorf("connect temporal: %w", err)
	}
	w := worker.New(c, taskQueue, worker.Options{})
	w.RegisterWorkflow(RunWorkflow)
	w.RegisterActivity(activities.ExecuteRunActivity)
	if err := w.Start(); err != nil {
		c.Close()
		return nil, fmt.Errorf("start temporal worker: %w", err)
	}
	return &TemporalRuntime{client: c, worker: w}, nil
}
func (r *TemporalRuntime) Close() { r.worker.Stop(); r.client.Close() }
