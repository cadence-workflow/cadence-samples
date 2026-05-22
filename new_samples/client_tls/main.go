package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/zap"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "start-workflow":
		startWorkflow()
	case "register-domain":
		registerDomain()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: go run . <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  start-workflow   Start a HelloWorld workflow with TLS")
	fmt.Println("  register-domain  Register a domain with TLS")
}

func startWorkflow() {
	logger := BuildLogger()
	logger.Info("Starting workflow with TLS...")

	tlsDialOption, err := BuildTLSDialOption()
	if err != nil {
		logger.Fatal("Failed to build TLS dial option", zap.Error(err))
	}

	cadenceClient := BuildCadenceClient(tlsDialOption)

	domain := "default"
	tasklist := "cadence-samples-worker"
	workflowID := uuid.New().String()
	requestID := uuid.New().String()
	executionTimeout := int32(60)
	closeTimeout := int32(60)

	workflowType := "cadence_samples.HelloWorldWorkflow"
	input := []byte(`{"message": "Cadence"}`)

	req := shared.StartWorkflowExecutionRequest{
		Domain:     &domain,
		WorkflowId: &workflowID,
		WorkflowType: &shared.WorkflowType{
			Name: &workflowType,
		},
		TaskList: &shared.TaskList{
			Name: &tasklist,
		},
		Input:                               input,
		ExecutionStartToCloseTimeoutSeconds: &executionTimeout,
		TaskStartToCloseTimeoutSeconds:      &closeTimeout,
		RequestId:                           &requestID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	resp, err := cadenceClient.StartWorkflowExecution(ctx, &req)
	if err != nil {
		logger.Fatal("Failed to start workflow", zap.Error(err))
	}

	logger.Info("Successfully started HelloWorld workflow",
		zap.String("workflowID", workflowID),
		zap.String("runID", resp.GetRunId()))
}

func registerDomain() {
	logger := BuildLogger()
	logger.Info("Registering domain with TLS...")

	tlsDialOption, err := BuildTLSDialOption()
	if err != nil {
		logger.Fatal("Failed to build TLS dial option", zap.Error(err))
	}

	cadenceClient := BuildCadenceClient(tlsDialOption)

	domain := "default"
	retentionDays := int32(7)
	emitMetric := true
	description := "Default domain for cadence samples"

	req := &shared.RegisterDomainRequest{
		Name:                                   &domain,
		Description:                            &description,
		WorkflowExecutionRetentionPeriodInDays: &retentionDays,
		EmitMetric:                             &emitMetric,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = cadenceClient.RegisterDomain(ctx, req)
	if err != nil {
		if _, ok := err.(*shared.DomainAlreadyExistsError); ok {
			logger.Info("Domain already exists", zap.String("domain", domain))
			return
		}
		logger.Fatal("Failed to register domain", zap.Error(err))
	}

	logger.Info("Successfully registered domain", zap.String("domain", domain))
}
