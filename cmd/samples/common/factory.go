package common

import (
	"errors"

	"github.com/opentracing/opentracing-go"
	"github.com/uber-go/tally"
	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/compatibility"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/workflow"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/zap"
)

const (
	_cadenceClientName      = "cadence-client"
	_cadenceFrontendService = "cadence-frontend"
)

// WorkflowClientBuilder build client to cadence service
type WorkflowClientBuilder struct {
	hostPort       string
	dispatcher     *yarpc.Dispatcher
	domain         string
	clientIdentity string
	metricsScope   tally.Scope
	Logger         *zap.Logger
	ctxProps       []workflow.ContextPropagator
	dataConverter  encoded.DataConverter
	tracer         opentracing.Tracer
}

// NewBuilder creates a new WorkflowClientBuilder
func NewBuilder(logger *zap.Logger) *WorkflowClientBuilder {
	return &WorkflowClientBuilder{
		Logger: logger,
	}
}

// SetHostPort sets the hostport for the builder
func (b *WorkflowClientBuilder) SetHostPort(hostport string) *WorkflowClientBuilder {
	b.hostPort = hostport
	return b
}

// SetDomain sets the domain for the builder
func (b *WorkflowClientBuilder) SetDomain(domain string) *WorkflowClientBuilder {
	b.domain = domain
	return b
}

// SetClientIdentity sets the identity for the builder
func (b *WorkflowClientBuilder) SetClientIdentity(identity string) *WorkflowClientBuilder {
	b.clientIdentity = identity
	return b
}

// SetMetricsScope sets the metrics scope for the builder
func (b *WorkflowClientBuilder) SetMetricsScope(metricsScope tally.Scope) *WorkflowClientBuilder {
	b.metricsScope = metricsScope
	return b
}

// SetDispatcher sets the dispatcher for the builder
func (b *WorkflowClientBuilder) SetDispatcher(dispatcher *yarpc.Dispatcher) *WorkflowClientBuilder {
	b.dispatcher = dispatcher
	return b
}

// SetContextPropagators sets the context propagators for the builder
func (b *WorkflowClientBuilder) SetContextPropagators(ctxProps []workflow.ContextPropagator) *WorkflowClientBuilder {
	b.ctxProps = ctxProps
	return b
}

// SetDataConverter sets the data converter for the builder
func (b *WorkflowClientBuilder) SetDataConverter(dataConverter encoded.DataConverter) *WorkflowClientBuilder {
	b.dataConverter = dataConverter
	return b
}

// SetTracer sets the tracer for the builder
func (b *WorkflowClientBuilder) SetTracer(tracer opentracing.Tracer) *WorkflowClientBuilder {
	b.tracer = tracer
	return b
}

// BuildCadenceClient builds a client to cadence service
func (b *WorkflowClientBuilder) BuildCadenceClient() (client.Client, error) {
	service, err := b.BuildServiceClient()
	if err != nil {
		return nil, err
	}

	return client.NewClient(
		service,
		b.domain,
		&client.Options{
			Identity:           b.clientIdentity,
			MetricsScope:       b.metricsScope,
			DataConverter:      b.dataConverter,
			ContextPropagators: b.ctxProps,
			Tracer:             b.tracer,
			FeatureFlags: client.FeatureFlags{
				WorkflowExecutionAlreadyCompletedErrorEnabled: true,
			},
		}), nil
}

// BuildCadenceDomainClient builds a domain client to cadence service
func (b *WorkflowClientBuilder) BuildCadenceDomainClient() (client.DomainClient, error) {
	service, err := b.BuildServiceClient()
	if err != nil {
		return nil, err
	}

	return client.NewDomainClient(
		service,
		&client.Options{
			Identity:           b.clientIdentity,
			MetricsScope:       b.metricsScope,
			ContextPropagators: b.ctxProps,
			FeatureFlags: client.FeatureFlags{
				WorkflowExecutionAlreadyCompletedErrorEnabled: true,
			},
		},
	), nil
}

// BuildServiceClient builds a rpc service client to cadence service
func (b *WorkflowClientBuilder) BuildServiceClient() (workflowserviceclient.Interface, error) {
	if err := b.build(); err != nil {
		return nil, err
	}

	if b.dispatcher == nil {
		b.Logger.Fatal("No RPC dispatcher provided to create a connection to Cadence Service")
	}

	clientConfig := b.dispatcher.ClientConfig(_cadenceFrontendService)
	return compatibility.NewThrift2ProtoAdapter(compatibility.AdapterClients{
		Domain:     apiv1.NewDomainAPIYARPCClient(clientConfig),
		Workflow:   apiv1.NewWorkflowAPIYARPCClient(clientConfig),
		Worker:     apiv1.NewWorkerAPIYARPCClient(clientConfig),
		Visibility: apiv1.NewVisibilityAPIYARPCClient(clientConfig),
		Schedule:   apiv1.NewScheduleAPIYARPCClient(clientConfig),
	}), nil
}

func (b *WorkflowClientBuilder) build() error {
	if b.dispatcher != nil {
		return nil
	}

	if len(b.hostPort) == 0 {
		return errors.New("HostPort is empty")
	}

	b.Logger.Debug("Creating RPC dispatcher outbound",
		zap.String("ServiceName", _cadenceFrontendService),
		zap.String("HostPort", b.hostPort))

	b.dispatcher = yarpc.NewDispatcher(yarpc.Config{
		Name: _cadenceClientName,
		Outbounds: yarpc.Outbounds{
			_cadenceFrontendService: {Unary: grpc.NewTransport().NewSingleOutbound(b.hostPort)},
		},
	})

	if b.dispatcher != nil {
		if err := b.dispatcher.Start(); err != nil {
			b.Logger.Fatal("Failed to create outbound transport channel: %v", zap.Error(err))
		}
	}

	return nil
}
