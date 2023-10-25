// Package worker implements a Cadence worker with basic configurations.
package worker

import (
	"github.com/uber-common/cadence-samples/new_samples/workflows"
	"github.com/uber-go/tally"
	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/compatibility"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	HostPort = "127.0.0.1:7833"
	Domain   = "cadence-samples"
	// TaskListName identifies set of client workflows, activities, and workers.
	// It could be your group or client or application name.
	TaskListName   = "cadence-samples-worker"
	ClientName     = "cadence-samples-worker"
	CadenceService = "cadence-frontend"
)

// StartWorker creates and starts a basic Cadence worker.
func StartWorker() {
	logger, cadenceClient := BuildLogger(), BuildCadenceClient()
	workerOptions := worker.Options{
		Logger:       logger,
		MetricsScope: tally.NewTestScope(TaskListName, nil),
	}

	w := worker.New(
		cadenceClient,
		Domain,
		TaskListName,
		workerOptions)
	// HelloWorld workflow registration
	w.RegisterWorkflowWithOptions(workflows.HelloWorldWorkflow, workflow.RegisterOptions{Name: "cadence_samples.HelloWorldWorkflow"})
	w.RegisterActivityWithOptions(workflows.HelloWorldActivity, activity.RegisterOptions{Name: "cadence_samples.HelloWorldActivity"})

	// Signal workflow registration
	w.RegisterWorkflowWithOptions(workflows.SignalGreeterMultiLanguageWorkflow, workflow.RegisterOptions{Name: "cadence_samples.SignalGreeterMultiLanguageWorkflow"})
	w.RegisterActivityWithOptions(workflows.GenerateGreetingMessage, activity.RegisterOptions{Name: "cadence_samples.GenerateGreetingMessage"})

	// Dynamic workflow registration
	w.RegisterWorkflowWithOptions(workflows.DynamicWorkflow, workflow.RegisterOptions{Name: "cadence_samples.DynamicWorkflow"})
	w.RegisterActivityWithOptions(workflows.DynamicGreetingActivity, activity.RegisterOptions{Name: "cadence_samples.DynamicGreetingActivity"})

	// ParallelPickFirst workflow registration
	w.RegisterWorkflowWithOptions(workflows.ParallelBranchPickFirstWorkflow, workflow.RegisterOptions{Name: "cadence_samples.ParallelBranchPickFirstWorkflow"})
	w.RegisterActivityWithOptions(workflows.ParallelActivity, activity.RegisterOptions{Name: "cadence_samples.ParallelActivity"})

	err := w.Start()
	if err != nil {
		panic("Failed to start worker: " + err.Error())
	}
	logger.Info("Started Worker.", zap.String("worker", TaskListName))

}

func BuildCadenceClient() workflowserviceclient.Interface {
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: ClientName,
		Outbounds: yarpc.Outbounds{
			CadenceService: {Unary: grpc.NewTransport().NewSingleOutbound(HostPort)},
		},
	})
	if err := dispatcher.Start(); err != nil {
		panic("Failed to start dispatcher: " + err.Error())
	}

	clientConfig := dispatcher.ClientConfig(CadenceService)

	return compatibility.NewThrift2ProtoAdapter(
		apiv1.NewDomainAPIYARPCClient(clientConfig),
		apiv1.NewWorkflowAPIYARPCClient(clientConfig),
		apiv1.NewWorkerAPIYARPCClient(clientConfig),
		apiv1.NewVisibilityAPIYARPCClient(clientConfig),
	)
}

func BuildLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zapcore.InfoLevel)

	var err error
	logger, err := config.Build()
	if err != nil {
		panic("Failed to setup logger: " + err.Error())
	}

	return logger
}
