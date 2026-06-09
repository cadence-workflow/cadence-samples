package main

import (
	"bytes"
	"context"
	"encoding/gob"

	"go.uber.org/cadence/client"
	"go.uber.org/cadence/encoded"
	"go.uber.org/zap"
)

// gobDataConverter is a custom DataConverter (gob instead of the default JSON) used to
// prove the ScheduleClient honors a caller-supplied converter when encoding Memo on write.
type gobDataConverter struct{}

func (gobDataConverter) ToData(values ...interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	for _, v := range values {
		if err := enc.Encode(v); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (gobDataConverter) FromData(input []byte, valuePtrs ...interface{}) error {
	dec := gob.NewDecoder(bytes.NewReader(input))
	for _, p := range valuePtrs {
		if err := dec.Decode(p); err != nil {
			return err
		}
	}
	return nil
}

// runDataConverter demonstrates the SDK's memo model end to end for both schedule-level
// and action-level Memo:
//
//	write: you pass native Go values; the client's DataConverter encodes them.
//	read : Describe returns raw bytes (map[string][]byte); you decode them yourself.
//
// We configure a custom gob converter to prove the SDK honors it on write for both memos,
// and that the default JSON converter cannot decode the gob bytes on read.
func runDataConverter() {
	logger := BuildLogger()
	c := buildScheduleClient(gobDataConverter{})
	sc := c.ScheduleClient()
	ctx := context.Background()
	scheduleID := newScheduleID("sample-dataconverter")
	defer deleteQuietly(logger, sc, context.Background(), scheduleID)

	const schedKey, schedVal = "team", "scheduling-team"
	const actionKey, actionVal = "runTeam", "execution-team"

	action := startWorkflowAction(logger, 0)
	action.Memo = map[string]interface{}{actionKey: actionVal}

	logger.Info("=== Create with schedule-level AND action-level Memo (encoded via custom gob converter) ===",
		zap.String("scheduleID", scheduleID))
	if _, err := sc.Create(ctx, &client.CreateScheduleRequest{
		ScheduleID: scheduleID,
		Spec:       &client.ScheduleSpec{CronExpression: "0 0 1 1 *"},
		Action:     &client.ScheduleAction{StartWorkflow: action},
		Memo:       map[string]interface{}{schedKey: schedVal},
	}); err != nil {
		logger.Fatal("Create failed", zap.Error(err))
	}

	desc, err := sc.Describe(ctx, scheduleID)
	if err != nil {
		logger.Fatal("Describe failed", zap.Error(err))
	}

	decodeMemoAndLog(logger, "schedule-level", desc.Memo, schedKey, schedVal)
	if desc.Action == nil || desc.Action.StartWorkflow == nil {
		logger.Fatal("Describe returned no action")
	}
	decodeMemoAndLog(logger, "action-level", desc.Action.StartWorkflow.Memo, actionKey, actionVal)

	// Proves the bytes were written with gob, not JSON.
	var wrong string
	if err = encoded.GetDefaultDataConverter().FromData(desc.Memo[schedKey], &wrong); err != nil {
		logger.Info("As expected, the default JSON converter CANNOT decode gob-encoded Memo",
			zap.String("error", err.Error()))
	} else {
		logger.Warn("Default converter unexpectedly decoded the Memo — converter may not have been honored")
	}
}

func decodeMemoAndLog(logger *zap.Logger, level string, memo map[string][]byte, key, want string) {
	raw, ok := memo[key]
	if !ok {
		logger.Fatal("Memo key missing from Describe response",
			zap.String("level", level), zap.String("key", key))
	}
	var got string
	if err := (gobDataConverter{}).FromData(raw, &got); err != nil {
		logger.Fatal("Failed to decode Memo with the custom converter",
			zap.String("level", level), zap.Error(err))
	}
	logger.Info("Decoded "+level+" Memo with the custom gob converter",
		zap.String("key", key), zap.String("value", got), zap.Bool("matches", got == want))
}
