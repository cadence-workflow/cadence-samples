# Data Converter Sample

This sample workflow demonstrates how to use custom data converters in Cadence workflows with compression capabilities. The data converter is responsible for serializing and deserializing workflow inputs, outputs, and activity parameters, with the added benefit of data compression to save storage space and bandwidth.

## Sample Description

The sample implements a custom compressed JSON data converter that:
- Serializes workflow inputs and activity parameters to JSON format
- Compresses the JSON data using gzip compression to reduce size
- Decompresses and deserializes workflow outputs and activity results from JSON format
- Provides significant storage and bandwidth savings for large payloads
- Demonstrates advanced data converter patterns for production use cases
- Shows real-time compression statistics and size comparisons

The sample includes two workflows:
1. **Simple Workflow**: Processes a basic `MyPayload` struct
2. **Large Payload Workflow**: Processes a complex `LargePayload` with nested objects, arrays, and extensive data to demonstrate compression benefits

All data is automatically compressed during serialization and decompressed during deserialization, with compression statistics displayed at runtime.

## Key Components

- **Custom Data Converter**: `compressedJSONDataConverter` implements the `encoded.DataConverter` interface with gzip compression
- **Simple Workflow**: `dataConverterWorkflow` demonstrates basic payload processing with compression
- **Large Payload Workflow**: `largeDataConverterWorkflow` demonstrates processing complex data structures with compression
- **Activities**: `dataConverterActivity` and `largeDataConverterActivity` process different payload types
- **Large Payload Generator**: `CreateLargePayload()` creates realistic complex data for compression demonstration
- **Compression Statistics**: `GetPayloadSizeInfo()` shows before/after compression metrics
- **Tests**: Includes unit tests for both simple and large payload workflows
- **Compression**: Automatic gzip compression/decompression for all workflow data

## Steps to Run Sample

1. You need a cadence service running. See details in cmd/samples/README.md

2. Run the following command to start the worker:
   ```
   ./bin/dataconverter -m worker
   ```

3. Run the following command to execute the workflow:
   ```
   ./bin/dataconverter -m trigger
   ```

You should see:
- Compression statistics showing original vs compressed data sizes
- Workflow logs showing the processing of large payloads
- Activity execution logs with payload information
- Final workflow completion with compression benefits noted

## Customization

To implement your own data converter with compression or other features:
1. Create a struct that implements the `encoded.DataConverter` interface
2. Implement the `ToData` method for serialization and compression
3. Implement the `FromData` method for decompression and deserialization
4. Register the converter in the worker options

This pattern is useful when you need to:
- Reduce storage costs and bandwidth usage with compression
- Use specific serialization formats for performance or compatibility
- Add encryption/decryption to workflow data
- Implement custom compression algorithms (LZ4, Snappy, etc.)
- Support legacy data formats
- Add data validation or transformation during serialization

## Performance Benefits

The compressed data converter provides:
- **Storage Savings**: Typically 60-80% reduction in data size for JSON payloads
- **Bandwidth Reduction**: Lower network transfer costs and faster data transmission
- **Cost Optimization**: Reduced storage costs in Cadence history
- **Scalability**: Better performance with large payloads 