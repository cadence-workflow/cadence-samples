package main

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// encryptedJSONDataConverter implements encoded.DataConverter with AES-256-GCM encryption.
// It serializes data to JSON, then encrypts using AES-256-GCM so that workflow history
// stored in Cadence is opaque to anyone without the key.
type encryptedJSONDataConverter struct {
	gcm cipher.AEAD
}

var errFailedToCreateConverter = errors.New("failed to create encrypted data converter")

// NewEncryptedJSONDataConverter creates a new encrypted JSON data converter.
// key must be exactly 32 bytes (AES-256). Returns an error if the key is invalid.
func NewEncryptedJSONDataConverter(key []byte) (encoded.DataConverter, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Join(errFailedToCreateConverter, err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Join(errFailedToCreateConverter, err)
	}
	return &encryptedJSONDataConverter{gcm: gcm}, nil
}

// demoEncryptionKey is a hardcoded 32-byte key used ONLY when CADENCE_ENCRYPTION_KEY is unset.
// DO NOT use this key in production. Rotate your key and load it from a secrets manager.
var demoEncryptionKey = []byte("cadence-demo-key-NOT-FOR-PROD!!!")

// LoadEncryptionKey reads a 32-byte AES key from the CADENCE_ENCRYPTION_KEY environment
// variable (hex-encoded, 64 hex chars). If the env var is unset, falls back to a hardcoded
// demo key with a warning. If the env var is set but invalid, panics — silently falling back
// to the public demo key when the user clearly intended their own key would be a security
// hole.
func LoadEncryptionKey() []byte {
	hexKey := os.Getenv("CADENCE_ENCRYPTION_KEY")
	if hexKey == "" {
		fmt.Println("WARNING: CADENCE_ENCRYPTION_KEY not set. Using hardcoded demo key.")
		fmt.Println("WARNING: DO NOT USE THE DEMO KEY IN PRODUCTION.")
		return demoEncryptionKey
	}
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		panic(fmt.Sprintf("CADENCE_ENCRYPTION_KEY is not valid hex: %v", err))
	}
	if len(key) != 32 {
		panic(fmt.Sprintf("CADENCE_ENCRYPTION_KEY must be exactly 64 hex chars (32 bytes), got %d hex chars (%d bytes)", len(hexKey), len(key)))
	}
	return key
}

func (dc *encryptedJSONDataConverter) ToData(value ...interface{}) ([]byte, error) {
	var jsonBuf bytes.Buffer
	enc := json.NewEncoder(&jsonBuf)
	for i, obj := range value {
		if err := enc.Encode(obj); err != nil {
			return nil, fmt.Errorf("unable to encode argument: %d, %v, with error: %v", i, reflect.TypeOf(obj), err)
		}
	}

	nonce := make([]byte, dc.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("unable to generate nonce: %v", err)
	}

	// Seal appends the GCM authentication tag to the ciphertext.
	// Output layout: nonce (12 bytes) || ciphertext+tag
	ciphertext := dc.gcm.Seal(nonce, nonce, jsonBuf.Bytes(), nil)
	return ciphertext, nil
}

func (dc *encryptedJSONDataConverter) FromData(input []byte, valuePtr ...interface{}) error {
	if len(input) == 0 {
		return nil
	}

	nonceSize := dc.gcm.NonceSize()
	if len(input) < nonceSize {
		return fmt.Errorf("ciphertext too short: %d bytes", len(input))
	}

	nonce, ciphertext := input[:nonceSize], input[nonceSize:]
	plaintext, err := dc.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return fmt.Errorf("decryption failed: %v", err)
	}

	dec := json.NewDecoder(bytes.NewBuffer(plaintext))
	for i, obj := range valuePtr {
		if err := dec.Decode(obj); err != nil {
			return fmt.Errorf("unable to decode argument: %d, %v, with error: %v", i, reflect.TypeOf(obj), err)
		}
	}
	return nil
}

// SensitiveCustomerRecord represents PII/PHI data that must be encrypted in workflow history.
type SensitiveCustomerRecord struct {
	CustomerID    string `json:"customer_id"`
	FullName      string `json:"full_name"`
	Email         string `json:"email"`
	SSN           string `json:"ssn"`
	CreditCard    string `json:"credit_card_number"`
	BillingAddr   string `json:"billing_address"`
	MedicalNotes  string `json:"medical_notes"`
	DiagnosisCode string `json:"diagnosis_code"`
	Prescriptions string `json:"prescriptions"`
	InsuranceID   string `json:"insurance_id"`
	ProcessedBy   string `json:"processed_by"`
}

// CreateSensitiveCustomerRecord creates a sample customer record with realistic PII/PHI.
func CreateSensitiveCustomerRecord() SensitiveCustomerRecord {
	return SensitiveCustomerRecord{
		CustomerID:    "cust_8a7f3b2e",
		FullName:      "Jane A. Doe",
		Email:         "jane.doe@example.com",
		SSN:           "123-45-6789",
		CreditCard:    "4111-1111-1111-1111",
		BillingAddr:   "1234 Elm Street, Springfield, IL 62701",
		MedicalNotes:  "Patient presents with hypertension and type-2 diabetes. Advised dietary changes and increased physical activity. Follow-up scheduled in 3 months.",
		DiagnosisCode: "I10, E11.9",
		Prescriptions: "Lisinopril 10mg once daily; Metformin 500mg twice daily",
		InsuranceID:   "INS-987654321",
		ProcessedBy:   "workflow-processor-v2",
	}
}

// GetEncryptionSizeInfo returns the plaintext size, ciphertext size, and a hex preview of the ciphertext.
func GetEncryptionSizeInfo(record SensitiveCustomerRecord, converter encoded.DataConverter) (int, int, string, error) {
	jsonData, err := json.Marshal(record)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to marshal record: %v", err)
	}
	plaintextSize := len(jsonData)

	encrypted, err := converter.ToData(record)
	if err != nil {
		return 0, 0, "", fmt.Errorf("failed to encrypt record: %v", err)
	}
	ciphertextSize := len(encrypted)

	preview := hex.EncodeToString(encrypted)
	if len(preview) > 80 {
		preview = preview[:80] + "..."
	}

	return plaintextSize, ciphertextSize, preview, nil
}

// EncryptionDataConverterWorkflow demonstrates encrypting sensitive workflow data.
// The DataConverter automatically encrypts all workflow inputs, outputs, and activity
// parameters before they are stored in Cadence history. Without the key, the data
// is unreadable even to Cadence operators viewing the workflow history.
//
// Note: The workflow generates its own payload internally so it can be started from
// the Cadence CLI without requiring the CLI to use the custom DataConverter.
func EncryptionDataConverterWorkflow(ctx workflow.Context) (SensitiveCustomerRecord, error) {
	logger := workflow.GetLogger(ctx)

	record := CreateSensitiveCustomerRecord()
	logger.Info("Encryption workflow started", zap.String("customer_id", record.CustomerID))
	logger.Info("All customer PII/PHI will be encrypted before storage in Cadence history")

	activityOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var result SensitiveCustomerRecord
	err := workflow.ExecuteActivity(ctx, EncryptionDataConverterActivity, record).Get(ctx, &result)
	if err != nil {
		logger.Error("Encryption workflow activity failed", zap.Error(err))
		return SensitiveCustomerRecord{}, err
	}

	logger.Info("Encryption workflow completed", zap.String("customer_id", result.CustomerID))
	logger.Info("Note: All PII/PHI was automatically encrypted/decrypted using AES-256-GCM")
	return result, nil
}

// EncryptionDataConverterActivity processes the sensitive customer record.
// In production this might perform claims processing, fraud checks, etc.
func EncryptionDataConverterActivity(ctx context.Context, record SensitiveCustomerRecord) (SensitiveCustomerRecord, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Encryption activity received record", zap.String("customer_id", record.CustomerID))

	record.ProcessedBy = record.ProcessedBy + " (Encrypted)"

	logger.Info("Encryption activity completed", zap.String("customer_id", record.CustomerID))
	return record, nil
}
