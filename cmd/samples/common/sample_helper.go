package common

import (
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/cadence/.gen/go/shared"

	prom "github.com/m3db/prometheus_client_golang/prometheus"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/prometheus"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFile = "config/development.yaml"
)

type (
	// SampleHelper class for workflow sample helper.
	SampleHelper struct {
		Service            workflowserviceclient.Interface
		WorkerMetricScope  tally.Scope
		ServiceMetricScope tally.Scope
		Logger             *zap.Logger
		Config             Configuration
		Builder            *WorkflowClientBuilder
		DataConverter      encoded.DataConverter
		CtxPropagators     []workflow.ContextPropagator
		workflowRegistries []registryOption
		activityRegistries []registryOption
		Tracer             opentracing.Tracer

		configFile string
	}

	// Configuration for running samples.
	Configuration struct {
		DomainName      string                    `yaml:"domain"`
		ServiceName     string                    `yaml:"service"`
		HostNameAndPort string                    `yaml:"host"`
		Prometheus      *prometheus.Configuration `yaml:"prometheus"`
	}

	registryOption struct {
		registry interface{}
		alias    string
	}
)

var (
	safeCharacters = []rune{'_'}

	sanitizeOptions = tally.SanitizeOptions{
		NameCharacters: tally.ValidCharacters{
			Ranges:     tally.AlphanumericRange,
			Characters: safeCharacters,
		},
		KeyCharacters: tally.ValidCharacters{
			Ranges:     tally.AlphanumericRange,
			Characters: safeCharacters,
		},
		ValueCharacters: tally.ValidCharacters{
			Ranges:     tally.AlphanumericRange,
			Characters: safeCharacters,
		},
		ReplacementCharacter: tally.DefaultReplacementCharacter,
	}
)

// SetConfigFile sets the config file path
func (h *SampleHelper) SetConfigFile(configFile string) {
	h.configFile = configFile
}

// SetupServiceConfig setup the config for the sample code run
func (h *SampleHelper) SetupServiceConfig() {
	if h.Service != nil {
		return
	}

	if h.configFile == "" {
		h.configFile = defaultConfigFile
	}
	// Initialize developer config for running samples
	configData, err := ioutil.ReadFile(h.configFile)
	if err != nil {
		panic(fmt.Sprintf("Failed to log config file: %v, Error: %v", defaultConfigFile, err))
	}

	if err := yaml.Unmarshal(configData, &h.Config); err != nil {
		panic(fmt.Sprintf("Error initializing configuration: %v", err))
	}

	// Initialize logger for running samples
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	logger.Info("Logger created.")
	h.Logger = logger
	h.ServiceMetricScope = tally.NoopScope
	h.WorkerMetricScope = tally.NoopScope

	if h.Config.Prometheus != nil {
		reporter, err := h.Config.Prometheus.NewReporter(
			prometheus.ConfigurationOptions{
				Registry: prom.NewRegistry(),
				OnError: func(err error) {
					logger.Warn("error in prometheus reporter", zap.Error(err))
				},
			},
		)
		if err != nil {
			panic(err)
		}

		h.WorkerMetricScope, _ = tally.NewRootScope(tally.ScopeOptions{
			Prefix:          "Worker_",
			Tags:            map[string]string{},
			CachedReporter:  reporter,
			Separator:       prometheus.DefaultSeparator,
			SanitizeOptions: &sanitizeOptions,
		}, 1*time.Second)

		// NOTE: this must be a different scope with different prefix, otherwise the metric will conflict
		h.ServiceMetricScope, _ = tally.NewRootScope(tally.ScopeOptions{
			Prefix:          "Service_",
			Tags:            map[string]string{},
			CachedReporter:  reporter,
			Separator:       prometheus.DefaultSeparator,
			SanitizeOptions: &sanitizeOptions,
		}, 1*time.Second)
	}
	h.Builder = NewBuilder(logger).
		SetHostPort(h.Config.HostNameAndPort).
		SetDomain(h.Config.DomainName).
		SetMetricsScope(h.ServiceMetricScope).
		SetDataConverter(h.DataConverter).
		SetTracer(h.Tracer).
		SetContextPropagators(h.CtxPropagators)
	service, err := h.Builder.BuildServiceClient()
	if err != nil {
		panic(err)
	}
	h.Service = service

	domainClient, _ := h.Builder.BuildCadenceDomainClient()
	_, err = domainClient.Describe(context.Background(), h.Config.DomainName)
	if err != nil {
		logger.Info("Domain doesn't exist", zap.String("Domain", h.Config.DomainName), zap.Error(err))
	} else {
		logger.Info("Domain successfully registered.", zap.String("Domain", h.Config.DomainName))
	}

	h.workflowRegistries = make([]registryOption, 0, 1)
	h.activityRegistries = make([]registryOption, 0, 1)
}

// StartWorkflow starts a workflow
func (h *SampleHelper) StartWorkflow(
	options client.StartWorkflowOptions,
	workflow interface{},
	args ...interface{},
) *workflow.Execution {
	return h.StartWorkflowWithCtx(context.Background(), options, workflow, args...)
}

// StartWorkflowWithCtx starts a workflow with the provided context
func (h *SampleHelper) StartWorkflowWithCtx(
	ctx context.Context,
	options client.StartWorkflowOptions,
	workflow interface{},
	args ...interface{},
) *workflow.Execution {
	workflowClient, err := h.Builder.BuildCadenceClient()
	if err != nil {
		h.Logger.Error("Failed to build cadence client.", zap.Error(err))
		panic(err)
	}

	we, err := workflowClient.StartWorkflow(ctx, options, workflow, args...)
	if err != nil {
		h.Logger.Error("Failed to create workflow", zap.Error(err))
		panic("Failed to create workflow.")
	} else {
		h.Logger.Info("Started Workflow", zap.String("WorkflowID", we.ID), zap.String("RunID", we.RunID))
		return we
	}
}

// SignalWithStartWorkflowWithCtx signals workflow and starts it if it's not yet started
func (h *SampleHelper) SignalWithStartWorkflowWithCtx(ctx context.Context, workflowID string, signalName string, signalArg interface{},
	options client.StartWorkflowOptions, workflow interface{}, workflowArgs ...interface{}) *workflow.Execution {
	workflowClient, err := h.Builder.BuildCadenceClient()
	if err != nil {
		h.Logger.Error("Failed to build cadence client.", zap.Error(err))
		panic(err)
	}

	we, err := workflowClient.SignalWithStartWorkflow(ctx, workflowID, signalName, signalArg, options, workflow, workflowArgs...)
	if err != nil {
		h.Logger.Error("Failed to signal with start workflow", zap.Error(err))
		panic("Failed to signal with start workflow.")

	} else {
		h.Logger.Info("Signaled and started Workflow", zap.String("WorkflowID", we.ID), zap.String("RunID", we.RunID))
	}
	return we
}

func (h *SampleHelper) RegisterWorkflow(workflow interface{}) {
	h.RegisterWorkflowWithAlias(workflow, "")
}

func (h *SampleHelper) RegisterWorkflowWithAlias(workflow interface{}, alias string) {
	registryOption := registryOption{
		registry: workflow,
		alias:    alias,
	}
	h.workflowRegistries = append(h.workflowRegistries, registryOption)
}

func (h *SampleHelper) RegisterActivity(activity interface{}) {
	h.RegisterActivityWithAlias(activity, "")
}

func (h *SampleHelper) RegisterActivityWithAlias(activity interface{}, alias string) {
	registryOption := registryOption{
		registry: activity,
		alias:    alias,
	}
	h.activityRegistries = append(h.activityRegistries, registryOption)
}

// StartWorkers starts workflow worker and activity worker based on configured options.
func (h *SampleHelper) StartWorkers(domainName string, groupName string, options worker.Options) worker.Worker {
	worker := worker.New(h.Service, domainName, groupName, options)
	h.registerWorkflowAndActivity(worker)

	err := worker.Start()
	if err != nil {
		h.Logger.Error("Failed to start workers.", zap.Error(err))
		panic("Failed to start workers")
	}

	return worker
}

func (h *SampleHelper) QueryWorkflow(workflowID, runID, queryType string, args ...interface{}) {
	workflowClient, err := h.Builder.BuildCadenceClient()
	if err != nil {
		h.Logger.Error("Failed to build cadence client.", zap.Error(err))
		panic(err)
	}

	resp, err := workflowClient.QueryWorkflow(context.Background(), workflowID, runID, queryType, args...)
	if err != nil {
		h.Logger.Error("Failed to query workflow", zap.Error(err))
		panic("Failed to query workflow.")
	}
	var result interface{}
	if err := resp.Get(&result); err != nil {
		h.Logger.Error("Failed to decode query result", zap.Error(err))
	}
	h.Logger.Info("Received query result", zap.Any("Result", result))
}

func (h *SampleHelper) ConsistentQueryWorkflow(
	valuePtr interface{},
	workflowID, runID, queryType string,
	args ...interface{},
) error {
	workflowClient, err := h.Builder.BuildCadenceClient()
	if err != nil {
		h.Logger.Error("Failed to build cadence client.", zap.Error(err))
		panic(err)
	}

	resp, err := workflowClient.QueryWorkflowWithOptions(context.Background(),
		&client.QueryWorkflowWithOptionsRequest{
			WorkflowID:            workflowID,
			RunID:                 runID,
			QueryType:             queryType,
			QueryConsistencyLevel: shared.QueryConsistencyLevelStrong.Ptr(),
			Args:                  args,
		})
	if err != nil {
		h.Logger.Error("Failed to query workflow", zap.Error(err))
		panic("Failed to query workflow.")
	}
	if err := resp.QueryResult.Get(&valuePtr); err != nil {
		h.Logger.Error("Failed to decode query result", zap.Error(err))
	}
	h.Logger.Info("Received consistent query result.", zap.Any("Result", valuePtr))
	return err
}

func (h *SampleHelper) SignalWorkflow(workflowID, signal string, data interface{}) {
	workflowClient, err := h.Builder.BuildCadenceClient()
	if err != nil {
		h.Logger.Error("Failed to build cadence client.", zap.Error(err))
		panic(err)
	}

	err = workflowClient.SignalWorkflow(context.Background(), workflowID, "", signal, data)
	if err != nil {
		h.Logger.Error("Failed to signal workflow", zap.Error(err))
		panic("Failed to signal workflow.")
	}
}

func (h *SampleHelper) CancelWorkflow(workflowID string) {
	workflowClient, err := h.Builder.BuildCadenceClient()
	if err != nil {
		h.Logger.Error("Failed to build cadence client.", zap.Error(err))
		panic(err)
	}

	err = workflowClient.CancelWorkflow(context.Background(), workflowID, "")
	if err != nil {
		h.Logger.Error("Failed to cancel workflow", zap.Error(err))
		panic("Failed to cancel workflow.")
	}
}

func (h *SampleHelper) registerWorkflowAndActivity(worker worker.Worker) {
	for _, w := range h.workflowRegistries {
		if len(w.alias) == 0 {
			worker.RegisterWorkflow(w.registry)
		} else {
			worker.RegisterWorkflowWithOptions(w.registry, workflow.RegisterOptions{Name: w.alias})
		}
	}
	for _, act := range h.activityRegistries {
		if len(act.alias) == 0 {
			worker.RegisterActivity(act.registry)
		} else {
			worker.RegisterActivityWithOptions(act.registry, activity.RegisterOptions{Name: act.alias})
		}
	}
}
