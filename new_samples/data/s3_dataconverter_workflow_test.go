package main

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/worker"
)

// memoryBlobStore is an in-memory BlobStore used in tests to avoid filesystem I/O.
type memoryBlobStore struct {
	mu   sync.RWMutex
	data map[string][]byte
}

func newMemoryBlobStore() BlobStore {
	return &memoryBlobStore{data: make(map[string][]byte)}
}

func (m *memoryBlobStore) Put(_ context.Context, key string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = append([]byte(nil), data...)
	return nil
}

func (m *memoryBlobStore) Get(_ context.Context, key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	d, ok := m.data[key]
	if !ok {
		return nil, nil
	}
	return append([]byte(nil), d...), nil
}

func Test_S3OffloadDataConverterWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(S3OffloadDataConverterWorkflow)
	env.RegisterActivity(S3OffloadDataConverterActivity)

	store := newMemoryBlobStore()
	dataConverter := NewS3OffloadDataConverter(store, "test-bucket", defaultThresholdBytes)
	workerOptions := worker.Options{
		DataConverter: dataConverter,
	}
	env.SetWorkerOptions(workerOptions)

	var activityResult S3LargePayload
	env.SetOnActivityCompletedListener(func(activityInfo *activity.Info, result encoded.Value, err error) {
		result.Get(&activityResult)
	})

	// Workflow generates its own payload internally, no input needed
	env.ExecuteWorkflow(S3OffloadDataConverterWorkflow)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, "batch-job-20240115-001", activityResult.JobID)
	require.Equal(t, "s3-offload-worker-v1 (Processed)", activityResult.ProcessedBy)
	require.Equal(t, 200, len(activityResult.DataPoints))
}

func Test_S3OffloadRoundTrip(t *testing.T) {
	store := newMemoryBlobStore()
	converter := NewS3OffloadDataConverter(store, "test-bucket", defaultThresholdBytes)

	original := CreateS3LargePayload()
	encoded, err := converter.ToData(original)
	require.NoError(t, err)
	require.NotEmpty(t, encoded)

	// Large payload should be offloaded — the encoded form should be tiny
	require.Equal(t, offloadPrefix, encoded[0], "expected offload prefix for large payload")
	require.Less(t, len(encoded), 200, "Cadence history reference should be much smaller than full payload")

	var decoded S3LargePayload
	err = converter.FromData(encoded, &decoded)
	require.NoError(t, err)
	require.Equal(t, original.JobID, decoded.JobID)
	require.Equal(t, len(original.DataPoints), len(decoded.DataPoints))
}

func Test_S3InlineSmallPayload(t *testing.T) {
	store := newMemoryBlobStore()
	converter := NewS3OffloadDataConverter(store, "test-bucket", 100000) // very high threshold

	original := CreateS3LargePayload()
	enc, err := converter.ToData(original)
	require.NoError(t, err)
	require.Equal(t, inlinePrefix, enc[0], "expected inline prefix when payload is under threshold")

	var decoded S3LargePayload
	err = converter.FromData(enc, &decoded)
	require.NoError(t, err)
	require.Equal(t, original.JobID, decoded.JobID)
}
