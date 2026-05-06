## Data Converter Samples

This folder demonstrates three production-ready patterns for custom `DataConverter` implementations in Cadence. A `DataConverter` controls how every workflow input, output, and activity parameter is serialized before it is written to Cadence history — making it the right place to add compression, encryption, or external offloading without changing any workflow or activity code.

### What is a DataConverter?

A `DataConverter` implements two methods:

- `ToData(value ...interface{}) ([]byte, error)` — called before data is written to Cadence history
- `FromData(input []byte, valuePtr ...interface{}) error` — called when data is read back

The same `DataConverter` must be used by **both the worker and any client that triggers or queries the workflow**. In this sample the workflows generate their own payloads internally, so they can be started from the Cadence CLI without bundling a custom converter into the CLI itself.

Each sample runs its own worker on its own task list so it can use its own `DataConverter`. Start all three with a single `go run .`.

---

### Compression Sample

`CompressionDataConverterWorkflow` demonstrates gzip-over-JSON compression. For repetitive JSON data this typically achieves 60–80% size reduction, lowering storage costs and bandwidth for large workflow payloads.

**Task list:** `cadence-samples-data-compression`

```bash
cadence --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.CompressionDataConverterWorkflow \
  --tl cadence-samples-data-compression \
  --et 60
```

When the worker starts it prints a compression statistics banner showing the before/after sizes of the sample payload so you can see the benefit immediately.

---

### Encryption Sample

`EncryptionDataConverterWorkflow` demonstrates AES-256-GCM encryption. Every workflow input, output, and activity parameter is encrypted before being written to Cadence history. Without the key, the data stored by the Cadence server — including any operators browsing workflow history — is completely opaque.

The sample uses a `SensitiveCustomerRecord` containing realistic PII and PHI fields (name, email, SSN, credit card, medical notes) to make the use case concrete.

**Task list:** `cadence-samples-data-encryption`

```bash
cadence --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.EncryptionDataConverterWorkflow \
  --tl cadence-samples-data-encryption \
  --et 60
```

#### Encryption key

By default, the worker uses a hardcoded demo key and prints a prominent warning. To use your own key:

```bash
# Generate a random 32-byte (256-bit) key
export CADENCE_ENCRYPTION_KEY=$(openssl rand -hex 32)
go run .
```

> **WARNING:** The hardcoded demo key (`cadence-demo-key-NOT-FOR-PROD!!!`) is public.
> Never use it in production. In production, load your key from a secrets manager
> (AWS Secrets Manager, HashiCorp Vault, GCP Secret Manager, etc.).

#### How AES-256-GCM works

- `ToData`: JSON-encode arguments → generate a 12-byte random nonce → `cipher.AEAD.Seal` → return `nonce || ciphertext+tag`.
- `FromData`: split nonce from input → `cipher.AEAD.Open` → JSON-decode.

The GCM authentication tag (16 bytes) ensures ciphertext tampering is detected. The random nonce means the same plaintext produces different ciphertext on every call, preventing replay detection by an attacker who observes Cadence history.

---

### S3 Offload Sample (Claim-Check Pattern)

`S3OffloadDataConverterWorkflow` demonstrates the *claim-check* pattern: payloads larger than a configurable threshold are stored in an external `BlobStore` and only a small reference (a few dozen bytes) travels through Cadence workflow history.

This solves the practical problem of Cadence's per-payload size limits (~2 MB) for workflows that must pass very large datasets between the workflow and its activities.

**Task list:** `cadence-samples-data-s3`

```bash
cadence --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.S3OffloadDataConverterWorkflow \
  --tl cadence-samples-data-s3 \
  --et 60
```

#### How it works

- `ToData`: JSON-encode → if `len(json) > thresholdBytes`, upload to `BlobStore` under a SHA-256 key and return `0x01 || {"__s3_ref":"<bucket>/<sha256hex>"}`. Otherwise return `0x00 || json` inline.
- `FromData`: read prefix byte → if `0x01`, fetch from `BlobStore` and decode; if `0x00`, decode inline.

#### Default store (zero-config)

Out of the box, `localFSBlobStore` writes blobs to `os.TempDir()/cadence-samples-data-s3/`. No cloud credentials or additional dependencies are needed.

#### Swapping in real AWS S3

The file `s3_dataconverter_workflow.go` contains a commented `s3BlobStore` stub showing the exact AWS SDK calls needed. To enable it:

1. Add the AWS SDK to your module:
   ```bash
   go get github.com/aws/aws-sdk-go-v2/config
   go get github.com/aws/aws-sdk-go-v2/service/s3
   ```
2. Uncomment the `s3BlobStore` section in `s3_dataconverter_workflow.go`.
3. Replace `NewLocalFSBlobStore()` with `NewS3BlobStore(bucket, region)` in `worker.go`.
4. Set standard AWS environment variables (`AWS_REGION`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`) or use an IAM instance role.

You can also point the SDK at [LocalStack](https://localstack.cloud/) or [MinIO](https://min.io/) for local testing without a real AWS account.

> **Note on cleanup:** The `s3OffloadDataConverter` does not delete blobs after the workflow completes. In production, use S3 object lifecycle policies to automatically expire old blobs.

---

### When to use which pattern

| Pattern | Best for |
|---------|----------|
| **Compression** | Large repetitive JSON payloads; reducing storage cost without confidentiality requirements |
| **Encryption** | PII, PHI, secrets, or any data that must be unreadable in Cadence history |
| **S3 Offload** | Payloads approaching Cadence's size limits; binary or non-JSON data; cost-conscious archival |

Patterns can be composed: encrypt-then-compress, or encrypt-then-offload to S3 for maximum security and minimum history size.
