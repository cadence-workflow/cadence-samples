package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// BlobStore is an abstraction over any external object store (local filesystem, S3, GCS, etc.).
// The s3OffloadDataConverter uses this interface to store large payloads outside Cadence history.
type BlobStore interface {
	Put(ctx context.Context, key string, data []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
}

// localFSBlobStore implements BlobStore using the local filesystem.
// It is the default zero-config implementation used when running this demo without real AWS.
// Files are written under os.TempDir()/cadence-samples-data-s3/.
type localFSBlobStore struct {
	baseDir string
}

// NewLocalFSBlobStore creates a local filesystem blob store under os.TempDir().
func NewLocalFSBlobStore() BlobStore {
	baseDir := filepath.Join(os.TempDir(), "cadence-samples-data-s3")
	if err := os.MkdirAll(baseDir, 0o755); err != nil {
		panic(fmt.Sprintf("failed to create blob store dir %s: %v", baseDir, err))
	}
	return &localFSBlobStore{baseDir: baseDir}
}

// sanitizeKey turns a "bucket/sha256hex" key into a single safe filename. Keys are always
// generated internally by the DataConverter, but filepath.Base provides a belt-and-suspenders
// guarantee against directory traversal in case a future caller passes a user-controlled key.
func sanitizeKey(key string) string {
	return filepath.Base(strings.ReplaceAll(key, "/", "_"))
}

func (s *localFSBlobStore) Put(_ context.Context, key string, data []byte) error {
	path := filepath.Join(s.baseDir, sanitizeKey(key))
	return os.WriteFile(path, data, 0o644)
}

func (s *localFSBlobStore) Get(_ context.Context, key string) ([]byte, error) {
	path := filepath.Join(s.baseDir, sanitizeKey(key))
	return os.ReadFile(path)
}

// =============================================================================
// S3 BlobStore stub
//
// To use a real AWS S3 bucket instead of the local filesystem:
//  1. Add aws-sdk-go-v2 to go.mod:
//       go get github.com/aws/aws-sdk-go-v2/config
//       go get github.com/aws/aws-sdk-go-v2/service/s3
//  2. Uncomment and compile the s3BlobStore implementation below.
//  3. Replace NewLocalFSBlobStore() with NewS3BlobStore(bucket, region) in worker.go.
//  4. Set AWS_REGION, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY (or use an instance role).
//
// /*
// import (
//     "github.com/aws/aws-sdk-go-v2/aws"
//     awsconfig "github.com/aws/aws-sdk-go-v2/config"
//     "github.com/aws/aws-sdk-go-v2/service/s3"
// )
//
// type s3BlobStore struct {
//     client *s3.Client
//     bucket string
// }
//
// func NewS3BlobStore(bucket, region string) BlobStore {
//     cfg, err := awsconfig.LoadDefaultConfig(context.Background(), awsconfig.WithRegion(region))
//     if err != nil {
//         panic("failed to load AWS config: " + err.Error())
//     }
//     return &s3BlobStore{client: s3.NewFromConfig(cfg), bucket: bucket}
// }
//
// func (s *s3BlobStore) Put(ctx context.Context, key string, data []byte) error {
//     _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
//         Bucket: aws.String(s.bucket),
//         Key:    aws.String(key),
//         Body:   bytes.NewReader(data),
//     })
//     return err
// }
//
// func (s *s3BlobStore) Get(ctx context.Context, key string) ([]byte, error) {
//     out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
//         Bucket: aws.String(s.bucket),
//         Key:    aws.String(key),
//     })
//     if err != nil {
//         return nil, err
//     }
//     defer out.Body.Close()
//     return io.ReadAll(out.Body)
// }
// */
// =============================================================================

// s3Envelope is the small reference stored in Cadence history when a payload is offloaded.
type s3Envelope struct {
	S3Ref string `json:"__s3_ref"`
}

const (
	// inlinePrefix is prepended to inline (below-threshold) payloads so FromData can distinguish them.
	inlinePrefix = byte(0x00)
	// offloadPrefix is prepended to offloaded payloads.
	offloadPrefix = byte(0x01)
	// defaultThresholdBytes: payloads larger than this are offloaded to the BlobStore.
	// Cadence's default max payload size is ~2MB; this threshold is set intentionally low
	// so the demo workflow comfortably triggers offloading.
	defaultThresholdBytes = 4096 // 4 KB
)

// s3OffloadDataConverter implements the claim-check pattern:
// large payloads are stored in BlobStore; only a small reference travels through Cadence history.
type s3OffloadDataConverter struct {
	store          BlobStore
	bucket         string
	thresholdBytes int
}

// NewS3OffloadDataConverter creates a new s3OffloadDataConverter.
// store is the BlobStore backend (use NewLocalFSBlobStore() for zero-config demo).
// bucket is a logical bucket/prefix name embedded in the reference key.
// thresholdBytes is the max inline payload size; larger payloads are offloaded.
func NewS3OffloadDataConverter(store BlobStore, bucket string, thresholdBytes int) encoded.DataConverter {
	return &s3OffloadDataConverter{
		store:          store,
		bucket:         bucket,
		thresholdBytes: thresholdBytes,
	}
}

func (dc *s3OffloadDataConverter) ToData(value ...interface{}) ([]byte, error) {
	var jsonBuf bytes.Buffer
	enc := json.NewEncoder(&jsonBuf)
	for i, obj := range value {
		if err := enc.Encode(obj); err != nil {
			return nil, fmt.Errorf("unable to encode argument: %d, %v, with error: %v", i, reflect.TypeOf(obj), err)
		}
	}
	jsonBytes := jsonBuf.Bytes()

	if len(jsonBytes) <= dc.thresholdBytes {
		// Small payload: store inline with a prefix marker
		result := make([]byte, 1+len(jsonBytes))
		result[0] = inlinePrefix
		copy(result[1:], jsonBytes)
		return result, nil
	}

	// Derive the key from the SHA-256 of the payload so ToData is idempotent across
	// Cadence workflow replays. Using uuid.New() here would write a new orphaned blob
	// on every replay because the SDK calls ToData again each time the workflow is
	// re-executed from the top. If the workflow needs to control the key (e.g. to
	// encode routing metadata), generate it with workflow.SideEffect and pass it
	// alongside the payload instead.
	hash := sha256.Sum256(jsonBytes)
	key := fmt.Sprintf("%s/%x", dc.bucket, hash)
	if err := dc.store.Put(context.Background(), key, jsonBytes); err != nil {
		return nil, fmt.Errorf("failed to offload payload to blob store (key=%s): %v", key, err)
	}

	envelope, err := json.Marshal(s3Envelope{S3Ref: key})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal s3 envelope: %v", err)
	}

	result := make([]byte, 1+len(envelope))
	result[0] = offloadPrefix
	copy(result[1:], envelope)
	return result, nil
}

func (dc *s3OffloadDataConverter) FromData(input []byte, valuePtr ...interface{}) error {
	// Empty input: workflow was started without arguments (e.g., from CLI without --input).
	if len(input) == 0 {
		return nil
	}

	prefix, payload := input[0], input[1:]

	var jsonData []byte
	switch prefix {
	case inlinePrefix:
		// Empty payload means zero arguments were encoded (e.g., no workflow input).
		if len(payload) == 0 {
			return nil
		}
		jsonData = payload
	case offloadPrefix:
		var envelope s3Envelope
		if err := json.Unmarshal(payload, &envelope); err != nil {
			return fmt.Errorf("s3 offload: failed to unmarshal envelope: %v", err)
		}
		fetched, err := dc.store.Get(context.Background(), envelope.S3Ref)
		if err != nil {
			return fmt.Errorf("s3 offload: failed to fetch payload from blob store (key=%s): %v", envelope.S3Ref, err)
		}
		jsonData = fetched
	default:
		return fmt.Errorf("s3 offload: unknown prefix byte 0x%02x", prefix)
	}

	dec := json.NewDecoder(bytes.NewBuffer(jsonData))
	for i, obj := range valuePtr {
		if err := dec.Decode(obj); err != nil {
			return fmt.Errorf("unable to decode argument: %d, %v, with error: %v", i, reflect.TypeOf(obj), err)
		}
	}
	return nil
}

// S3LargePayload is a sizable data structure used to demonstrate S3 offloading.
// It is intentionally larger than defaultThresholdBytes so every workflow execution
// triggers an offload to the BlobStore.
type S3LargePayload struct {
	JobID       string            `json:"job_id"`
	Description string            `json:"description"`
	DataPoints  []S3DataPoint     `json:"data_points"`
	Metadata    map[string]string `json:"metadata"`
	ProcessedBy string            `json:"processed_by"`
}

// S3DataPoint represents a single telemetry measurement.
type S3DataPoint struct {
	Timestamp string  `json:"timestamp"`
	Metric    string  `json:"metric"`
	Value     float64 `json:"value"`
	Tags      string  `json:"tags"`
}

// CreateS3LargePayload creates a sample payload well above defaultThresholdBytes.
func CreateS3LargePayload() S3LargePayload {
	points := make([]S3DataPoint, 200)
	for i := range points {
		points[i] = S3DataPoint{
			Timestamp: fmt.Sprintf("2024-01-15T%02d:30:00Z", i%24),
			Metric:    fmt.Sprintf("telemetry.sensor_%03d.temperature", i),
			Value:     20.0 + float64(i%30)/10.0,
			Tags:      fmt.Sprintf("region=us-east-1,host=node-%03d,env=production", i%10),
		}
	}

	meta := make(map[string]string)
	for i := 0; i < 20; i++ {
		meta[fmt.Sprintf("batch_key_%02d", i)] = strings.Repeat("value-data-", 5)
	}

	return S3LargePayload{
		JobID:       "batch-job-20240115-001",
		Description: strings.Repeat("Large telemetry batch job containing sensor readings from the production cluster. ", 10),
		DataPoints:  points,
		Metadata:    meta,
		ProcessedBy: "s3-offload-worker-v1",
	}
}

// GetS3OffloadSizeInfo returns the JSON size, what is stored externally, and what travels through Cadence.
func GetS3OffloadSizeInfo(payload S3LargePayload, thresholdBytes int) (int, int, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to marshal payload: %v", err)
	}
	jsonSize := len(jsonData)

	// The Cadence history reference is: 1 prefix byte + JSON envelope {"__s3_ref":"<bucket>/<sha256hex>"}
	// A SHA-256 hex digest is 64 chars; bucket + "/" + hex ≈ bucket + 65 chars
	sampleEnvelope, _ := json.Marshal(s3Envelope{S3Ref: "cadence-samples-data-s3/" + strings.Repeat("a", 64)})
	cadenceBytes := 1 + len(sampleEnvelope)

	return jsonSize, cadenceBytes, nil
}

// S3OffloadDataConverterWorkflow demonstrates the claim-check pattern with a BlobStore.
// Payloads larger than the threshold are stored externally; only a small reference is
// kept in Cadence workflow history, dramatically reducing history storage requirements.
//
// Note: The workflow generates its own payload internally so it can be started from
// the Cadence CLI without requiring the CLI to use the custom DataConverter.
func S3OffloadDataConverterWorkflow(ctx workflow.Context) (S3LargePayload, error) {
	logger := workflow.GetLogger(ctx)

	payload := CreateS3LargePayload()
	logger.Info("S3 offload workflow started",
		zap.String("job_id", payload.JobID),
		zap.Int("data_points", len(payload.DataPoints)))
	logger.Info("Large payload will be offloaded to BlobStore; only a reference travels through Cadence history")

	activityOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var result S3LargePayload
	err := workflow.ExecuteActivity(ctx, S3OffloadDataConverterActivity, payload).Get(ctx, &result)
	if err != nil {
		logger.Error("S3 offload workflow activity failed", zap.Error(err))
		return S3LargePayload{}, err
	}

	logger.Info("S3 offload workflow completed", zap.String("job_id", result.JobID))
	logger.Info("Note: Large payload was transparently offloaded and retrieved via the BlobStore")
	return result, nil
}

// S3OffloadDataConverterActivity processes the large payload retrieved from the BlobStore.
// From the activity's perspective the DataConverter is invisible — it receives the full
// deserialized struct just as it would with any other DataConverter.
func S3OffloadDataConverterActivity(ctx context.Context, payload S3LargePayload) (S3LargePayload, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("S3 offload activity received payload",
		zap.String("job_id", payload.JobID),
		zap.Int("data_points", len(payload.DataPoints)))

	payload.ProcessedBy = payload.ProcessedBy + " (Processed)"

	logger.Info("S3 offload activity completed", zap.String("job_id", payload.JobID))
	return payload, nil
}
