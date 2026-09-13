package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCExecutor struct{ conn *grpc.ClientConn }

func NewGRPCExecutor(addr string) (*GRPCExecutor, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.ForceCodec(protoWireCodec{})))
	if err != nil {
		return nil, fmt.Errorf("connect to worker: %w", err)
	}
	return &GRPCExecutor{conn: conn}, nil
}
func (e *GRPCExecutor) Close() error { return e.conn.Close() }
func (e *GRPCExecutor) Execute(ctx context.Context, run Run, input string) (string, error) {
	var response ExecuteRunResponse
	err := e.conn.Invoke(ctx, "/agentmesh.run.v1.RunWorker/ExecuteRun", &ExecuteRunRequest{RunID: run.ID, AgentID: run.AgentID, Input: input}, &response, grpc.ForceCodec(protoWireCodec{}))
	if err != nil {
		return "", fmt.Errorf("worker ExecuteRun: %w", err)
	}
	if response.Status != "completed" {
		return "", fmt.Errorf("worker returned status %q", response.Status)
	}
	return response.Result, nil
}
