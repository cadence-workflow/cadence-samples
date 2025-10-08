package common

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"

	"github.com/opentracing/opentracing-go"
	"github.com/uber-go/tally"
	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/compatibility"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/workflow"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/peer"
	"go.uber.org/yarpc/peer/hostport"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
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
	tlsConfig      *tls.Config
	clientCertPath string
	clientKeyPath  string
	caCertPath     string
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

// SetTLSConfig sets the TLS configuration for the builder
func (b *WorkflowClientBuilder) SetTLSConfig(tlsConfig *tls.Config) *WorkflowClientBuilder {
	b.tlsConfig = tlsConfig
	return b
}

// SetTLSCertificates sets the TLS certificate paths for the builder
func (b *WorkflowClientBuilder) SetTLSCertificates(clientCertPath, clientKeyPath, caCertPath string) *WorkflowClientBuilder {
	b.clientCertPath = clientCertPath
	b.clientKeyPath = clientKeyPath
	b.caCertPath = caCertPath
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
	return compatibility.NewThrift2ProtoAdapter(
		apiv1.NewDomainAPIYARPCClient(clientConfig),
		apiv1.NewWorkflowAPIYARPCClient(clientConfig),
		apiv1.NewWorkerAPIYARPCClient(clientConfig),
		apiv1.NewVisibilityAPIYARPCClient(clientConfig),
	), nil
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

	// Check if TLS is configured
	if b.tlsConfig != nil || (b.clientCertPath != "" && b.clientKeyPath != "" && b.caCertPath != "") {
		// Build TLS configuration if certificate paths are provided but tlsConfig is not
		if b.tlsConfig == nil {
			tlsConfig, err := b.buildTLSConfig()
			if err != nil {
				return err
			}
			b.tlsConfig = tlsConfig
		}

		// Create TLS-enabled gRPC transport
		grpcTransport := grpc.NewTransport()
		var dialOptions []grpc.DialOption

		creds := credentials.NewTLS(b.tlsConfig)
		dialOptions = append(dialOptions, grpc.DialerCredentials(creds))

		dialer := grpcTransport.NewDialer(dialOptions...)
		outbound := grpcTransport.NewOutbound(peer.NewSingle(hostport.PeerIdentifier(b.hostPort), dialer))

		b.dispatcher = yarpc.NewDispatcher(yarpc.Config{
			Name: _cadenceClientName,
			Outbounds: yarpc.Outbounds{
				_cadenceFrontendService: {Unary: outbound},
			},
		})
	} else {
		// Create standard non-TLS dispatcher
		b.dispatcher = yarpc.NewDispatcher(yarpc.Config{
			Name: _cadenceClientName,
			Outbounds: yarpc.Outbounds{
				_cadenceFrontendService: {Unary: grpc.NewTransport().NewSingleOutbound(b.hostPort)},
			},
		})
	}

	if b.dispatcher != nil {
		if err := b.dispatcher.Start(); err != nil {
			b.Logger.Fatal("Failed to create outbound transport channel: %v", zap.Error(err))
		}
	}

	return nil
}

// buildTLSConfig creates a TLS configuration from certificate paths
func (b *WorkflowClientBuilder) buildTLSConfig() (*tls.Config, error) {
	// Present client cert for mutual TLS (if enabled on server)
	clientCert, err := tls.LoadX509KeyPair(b.clientCertPath, b.clientKeyPath)
	if err != nil {
		b.Logger.Fatal("Failed to load client certificate: %v", zap.Error(err))
		return nil, err
	}

	// Load server CA
	caCert, err := ioutil.ReadFile(b.caCertPath)
	if err != nil {
		b.Logger.Fatal("Failed to load server CA certificate: %v", zap.Error(err))
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{clientCert},
	}

	return tlsConfig, nil
}
