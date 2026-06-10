// THIS IS A GENERATED FILE
// PLEASE DO NOT EDIT

// Package worker implements a Cadence worker with basic configurations.
package main

import (
	"github.com/uber-go/tally"
	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/compatibility"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/peer"
	yarpchostport "go.uber.org/yarpc/peer/hostport"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	HostPort       = "127.0.0.1:7833"
	Domain         = "cadence-samples"
	TaskListName   = "schedule-sample-worker"
	ClientName     = "schedule-sample-worker"
	CadenceService = "cadence-frontend"
)

// StartWorker creates and starts a Cadence worker for the schedule sample.
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
	w.RegisterWorkflowWithOptions(scheduledWorkflow, workflow.RegisterOptions{Name: scheduledWorkflowName})
	w.RegisterActivityWithOptions(scheduledActivity, activity.RegisterOptions{Name: "scheduledActivity"})

	err := w.Start()
	if err != nil {
		panic("Failed to start worker: " + err.Error())
	}
	logger.Info("Started Worker.", zap.String("worker", TaskListName))
}

func BuildCadenceClient(dialOptions ...grpc.DialOption) workflowserviceclient.Interface {
	grpcTransport := grpc.NewTransport()
	myChooser := peer.NewSingle(
		yarpchostport.Identify(HostPort),
		grpcTransport.NewDialer(dialOptions...),
	)
	outbound := grpcTransport.NewOutbound(myChooser)

	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: ClientName,
		Outbounds: yarpc.Outbounds{
			CadenceService: {Unary: outbound},
		},
	})
	if err := dispatcher.Start(); err != nil {
		panic("Failed to start dispatcher: " + err.Error())
	}

	clientConfig := dispatcher.ClientConfig(CadenceService)

	return compatibility.NewThrift2ProtoAdapter(compatibility.AdapterClients{
		Domain:     apiv1.NewDomainAPIYARPCClient(clientConfig),
		Workflow:   apiv1.NewWorkflowAPIYARPCClient(clientConfig),
		Worker:     apiv1.NewWorkerAPIYARPCClient(clientConfig),
		Visibility: apiv1.NewVisibilityAPIYARPCClient(clientConfig),
		Schedule:   apiv1.NewScheduleAPIYARPCClient(clientConfig),
	})
}

func BuildLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zapcore.InfoLevel)

	logger, err := config.Build()
	if err != nil {
		panic("Failed to setup logger: " + err.Error())
	}
	return logger
}
