package main

import (
	"bytes"
	"context"
	"encoding/gob"

	"go.uber.org/cadence/client"
	"go.uber.org/cadence/encoded"
	"go.uber.org/zap"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
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

// runDataConverter demonstrates the SDK's memo model end to end, for BOTH schedule-level
// and action-level Memo:
//
//	write: you pass native Go values; the client's DataConverter encodes them.
//	read : Describe returns raw bytes (map[string][]byte); you decode them yourself.
//
// We configure a custom gob converter to prove (a) the SDK honors it on write for both
// memos, and (b) the default JSON converter cannot decode the gob bytes on read.
//
// No worker is required: both memos round-trip through Create → Describe directly.
func runDataConverter(h *common.SampleHelper) {
	// Build a client whose ScheduleClient uses our custom gob converter.
	c, err := h.Builder.SetDataConverter(gobDataConverter{}).BuildCadenceClient()
	if err != nil {
		h.Logger.Fatal("Failed to build client with custom DataConverter", zap.Error(err))
	}
	sc := c.ScheduleClient()
	ctx := context.Background()
	scheduleID := newScheduleID("sample-dataconverter")
	defer deleteQuietly(h, sc, context.Background(), scheduleID)

	const schedKey, schedVal = "team", "scheduling-team"
	const actionKey, actionVal = "runTeam", "execution-team"

	// Both memos are written as native Go values — the SDK encodes each via the gob converter.
	action := startWorkflowAction(h, 0)
	action.Memo = map[string]interface{}{actionKey: actionVal}

	h.Logger.Info("=== Create with schedule-level AND action-level Memo (encoded via custom gob converter) ===",
		zap.String("scheduleID", scheduleID))
	if _, err = sc.Create(ctx, &client.CreateScheduleRequest{
		ScheduleID: scheduleID,
		Spec:       &client.ScheduleSpec{CronExpression: "0 0 1 1 *"}, // yearly: won't fire during demo
		Action:     &client.ScheduleAction{StartWorkflow: action},
		Memo:       map[string]interface{}{schedKey: schedVal},
	}); err != nil {
		h.Logger.Fatal("Create failed", zap.Error(err))
	}

	desc, err := sc.Describe(ctx, scheduleID)
	if err != nil {
		h.Logger.Fatal("Describe failed", zap.Error(err))
	}

	// Both memos come back as raw bytes; decode each with the same gob converter.
	decodeMemoAndLog(h, "schedule-level", desc.Memo, schedKey, schedVal)
	if desc.Action == nil || desc.Action.StartWorkflow == nil {
		h.Logger.Fatal("Describe returned no action")
	}
	decodeMemoAndLog(h, "action-level", desc.Action.StartWorkflow.Memo, actionKey, actionVal)

	// Decoding the gob bytes with the DEFAULT (JSON) converter fails — proving the bytes
	// were written with our custom converter, not hard-coded JSON.
	var wrong string
	if err = encoded.GetDefaultDataConverter().FromData(desc.Memo[schedKey], &wrong); err != nil {
		h.Logger.Info("As expected, the default JSON converter CANNOT decode gob-encoded Memo",
			zap.String("error", err.Error()))
	} else {
		h.Logger.Warn("Default converter unexpectedly decoded the Memo — converter may not have been honored")
	}
}

// decodeMemoAndLog decodes one raw memo field with the gob converter and logs the round-trip.
func decodeMemoAndLog(h *common.SampleHelper, level string, memo map[string][]byte, key, want string) {
	raw, ok := memo[key]
	if !ok {
		h.Logger.Fatal("Memo key missing from Describe response",
			zap.String("level", level), zap.String("key", key))
	}
	var got string
	if err := (gobDataConverter{}).FromData(raw, &got); err != nil {
		h.Logger.Fatal("Failed to decode Memo with the custom converter",
			zap.String("level", level), zap.Error(err))
	}
	h.Logger.Info("Decoded "+level+" Memo with the custom gob converter",
		zap.String("key", key), zap.String("value", got), zap.Bool("matches", got == want))
}
