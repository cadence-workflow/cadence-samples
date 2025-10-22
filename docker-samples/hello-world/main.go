package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/compatibility"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/grpc"
)

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var (
	// HostPort - Cadence server host:port (configurable via CADENCE_HOST env var)
	HostPort = getEnv("CADENCE_HOST", "127.0.0.1:7833")
	// Domain - Cadence domain name (configurable via CADENCE_DOMAIN env var)
	Domain = getEnv("CADENCE_DOMAIN", "test-domain")
	// TaskListName - Task list for this worker (configurable via TASK_LIST env var)
	TaskListName = getEnv("TASK_LIST", "test-worker")
	// ClientName - YARPC client name
	ClientName = getEnv("CLIENT_NAME", "test-worker")
	// CadenceService - Service name for YARPC routing
	CadenceService = "cadence-frontend"
)

func main() {
	startWorker(buildLogger(), buildCadenceClient())
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func buildLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zapcore.InfoLevel)

	var err error
	logger, err := config.Build()
	if err != nil {
		panic("Failed to setup logger")
	}

	return logger
}

func buildCadenceClient() workflowserviceclient.Interface {
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: ClientName,
		Outbounds: yarpc.Outbounds{
			CadenceService: {Unary: grpc.NewTransport().NewSingleOutbound(HostPort)},
		},
	})
	if err := dispatcher.Start(); err != nil {
		panic("Failed to start dispatcher")
	}

	clientConfig := dispatcher.ClientConfig(CadenceService)

	return compatibility.NewThrift2ProtoAdapter(
		apiv1.NewDomainAPIYARPCClient(clientConfig),
		apiv1.NewWorkflowAPIYARPCClient(clientConfig),
		apiv1.NewWorkerAPIYARPCClient(clientConfig),
		apiv1.NewVisibilityAPIYARPCClient(clientConfig),
	)
}

func startWorker(logger *zap.Logger, service workflowserviceclient.Interface) {
	// TaskListName identifies set of client workflows, activities, and workers.
	// It could be your group or client or application name.
	workerOptions := worker.Options{
		Logger:       logger,
		MetricsScope: tally.NewTestScope(TaskListName, map[string]string{}),
	}

	worker, err := worker.NewV2(
		service,
		Domain,
		TaskListName,
		workerOptions)
	if err != nil {
		panic("Failed to initialize worker")
	}

	// Register workflow and activity
	worker.RegisterWorkflow(helloWorldWorkflow)
	worker.RegisterActivity(helloWorldActivity)

	err = worker.Start()
	if err != nil {
		panic("Failed to start worker")
	}

	logger.Info("Started Worker.", zap.String("worker", TaskListName))
}

func helloWorldWorkflow(ctx workflow.Context, name string) (*string, error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("helloworld workflow started")
	var helloworldResult string
	err := workflow.ExecuteActivity(ctx, helloWorldActivity, name).Get(ctx, &helloworldResult)
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return nil, err
	}

	logger.Info("Workflow completed.", zap.String("Result", helloworldResult))

	return &helloworldResult, nil
}

func helloWorldActivity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("helloworld activity started")
	return "Hello " + name + "!", nil
}

