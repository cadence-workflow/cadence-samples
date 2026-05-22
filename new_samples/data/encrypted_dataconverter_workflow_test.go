package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/worker"
)

func Test_EncryptionDataConverterWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(EncryptionDataConverterWorkflow)
	env.RegisterActivity(EncryptionDataConverterActivity)

	dataConverter, err := NewEncryptedJSONDataConverter(demoEncryptionKey)
	require.NoError(t, err)
	workerOptions := worker.Options{
		DataConverter: dataConverter,
	}
	env.SetWorkerOptions(workerOptions)

	var activityResult SensitiveCustomerRecord
	env.SetOnActivityCompletedListener(func(activityInfo *activity.Info, result encoded.Value, err error) {
		result.Get(&activityResult)
	})

	// Workflow generates its own payload internally, no input needed
	env.ExecuteWorkflow(EncryptionDataConverterWorkflow)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, "cust_8a7f3b2e", activityResult.CustomerID)
	require.Equal(t, "workflow-processor-v2 (Encrypted)", activityResult.ProcessedBy)
}

func Test_EncryptionRoundTrip(t *testing.T) {
	converter, err := NewEncryptedJSONDataConverter(demoEncryptionKey)
	require.NoError(t, err)

	original := CreateSensitiveCustomerRecord()
	encrypted, err := converter.ToData(original)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	var decoded SensitiveCustomerRecord
	err = converter.FromData(encrypted, &decoded)
	require.NoError(t, err)
	require.Equal(t, original.SSN, decoded.SSN)
	require.Equal(t, original.CreditCard, decoded.CreditCard)
	require.Equal(t, original.MedicalNotes, decoded.MedicalNotes)
}

func Test_EncryptionDifferentEachTime(t *testing.T) {
	converter, err := NewEncryptedJSONDataConverter(demoEncryptionKey)
	require.NoError(t, err)
	record := CreateSensitiveCustomerRecord()

	enc1, err := converter.ToData(record)
	require.NoError(t, err)
	enc2, err := converter.ToData(record)
	require.NoError(t, err)

	// Each encryption produces a different ciphertext due to random nonce
	require.NotEqual(t, enc1, enc2)
}

func Test_NewEncryptedJSONDataConverter_InvalidKey(t *testing.T) {
	_, err := NewEncryptedJSONDataConverter([]byte("too-short"))
	require.Error(t, err)
	require.ErrorIs(t, err, errFailedToCreateConverter)
}
