// Package common provides shared utilities for Cadence samples.
// This simplifies worker setup by providing a one-liner for client creation.
//
// Once cadence-client adds NewGrpcClient(), this package can be replaced
// with a direct import from go.uber.org/cadence/client.
package common

import (
	"fmt"

	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/compatibility"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/peer"
	"go.uber.org/yarpc/peer/hostport"
	"go.uber.org/yarpc/transport/grpc"
)

const (
	// DefaultHostPort is the default Cadence frontend address
	DefaultHostPort = "127.0.0.1:7833"
	// DefaultDomain is the default domain for samples
	DefaultDomain = "cadence-samples"
	// DefaultTaskList is the default task list for samples
	DefaultTaskList = "cadence-samples-worker"
	// CadenceService is the service name for YARPC
	CadenceService = "cadence-frontend"
)

// NewCadenceClient creates a new Cadence client connected via gRPC.
// This is a simplified helper that handles all the YARPC/gRPC boilerplate.
//
// Parameters:
//   - caller: The name of the calling service (used for tracing)
//   - hostPort: The Cadence frontend address (e.g., "localhost:7833")
//   - dialOptions: Optional gRPC dial options (e.g., for TLS)
//
// Example:
//
//	client, err := common.NewCadenceClient("my-worker", "localhost:7833")
func NewCadenceClient(caller, hostPort string, dialOptions ...grpc.DialOption) (workflowserviceclient.Interface, error) {
	grpcTransport := grpc.NewTransport()

	myChooser := peer.NewSingle(
		hostport.Identify(hostPort),
		grpcTransport.NewDialer(dialOptions...),
	)
	outbound := grpcTransport.NewOutbound(myChooser)

	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: caller,
		Outbounds: yarpc.Outbounds{
			CadenceService: {Unary: outbound},
		},
	})
	if err := dispatcher.Start(); err != nil {
		return nil, fmt.Errorf("failed to start dispatcher: %w", err)
	}

	clientConfig := dispatcher.ClientConfig(CadenceService)

	return compatibility.NewThrift2ProtoAdapter(
		apiv1.NewDomainAPIYARPCClient(clientConfig),
		apiv1.NewWorkflowAPIYARPCClient(clientConfig),
		apiv1.NewWorkerAPIYARPCClient(clientConfig),
		apiv1.NewVisibilityAPIYARPCClient(clientConfig),
	), nil
}

// MustNewCadenceClient is like NewCadenceClient but panics on error.
// Useful for sample code where error handling would add noise.
func MustNewCadenceClient(caller, hostPort string, dialOptions ...grpc.DialOption) workflowserviceclient.Interface {
	client, err := NewCadenceClient(caller, hostPort, dialOptions...)
	if err != nil {
		panic(err)
	}
	return client
}
