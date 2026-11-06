package main

import (
	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/compatibility"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/grpc"
)

const (
	Domain         = "cadence-samples-active-active"
	TaskListName   = "cadence-samples-worker"
	CadenceService = "cadence-frontend"
)

// Clusters maps each cluster to its frontend address.
var Clusters = map[string]string{
	"cluster0": "127.0.0.1:7833",
	"cluster1": "127.0.0.1:7843",
}

// RunWorker runs a worker that polls the given cluster until it is interrupted.
func RunWorker(cluster string) error {
	w, err := worker.NewV2(BuildCadenceClient(Clusters[cluster]), Domain, TaskListName, worker.Options{})
	if err != nil {
		return err
	}
	w.RegisterWorkflowWithOptions(ActiveActiveWorkflow, workflow.RegisterOptions{Name: ActiveActiveWorkflowName})
	w.RegisterActivityWithOptions((&ClusterActivities{Cluster: cluster}).GetCluster, activity.RegisterOptions{Name: GetClusterActivityName})

	return w.Run()
}

func BuildCadenceClient(hostPort string) workflowserviceclient.Interface {
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: TaskListName,
		Outbounds: yarpc.Outbounds{
			CadenceService: {Unary: grpc.NewTransport().NewSingleOutbound(hostPort)},
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
	})
}
