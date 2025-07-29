# Data Converter Sample

This sample workflow demonstrates how to use custom data converters in Cadence workflows. The data converter is responsible for serializing and deserializing workflow inputs, outputs, and activity parameters.

## Sample Description

The sample implements a custom JSON data converter that:
- Serializes workflow inputs and activity parameters to JSON format
- Deserializes workflow outputs and activity results from JSON format
- Provides better control over data serialization compared to the default data converter
- Can be extended to support custom serialization formats (e.g., Protocol Buffers, MessagePack)

The workflow takes a `MyPayload` struct as input, processes it through an activity, and returns the modified payload.

## Key Components

- **Custom Data Converter**: `jsonDataConverter` implements the `encoded.DataConverter` interface
- **Workflow**: `dataConverterWorkflow` demonstrates using custom data types with the converter
- **Activity**: `dataConverterActivity` processes the input and returns modified data
- **Test**: Includes unit tests to verify the data converter functionality

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

You should see logs showing the workflow input being processed through the activity and the final result being returned.

## Customization

To use a different serialization format, you can implement your own data converter by:
1. Creating a struct that implements the `encoded.DataConverter` interface
2. Implementing the `ToData` method for serialization
3. Implementing the `FromData` method for deserialization
4. Registering the converter in the worker options

This pattern is useful when you need to:
- Use specific serialization formats for performance or compatibility
- Add encryption/decryption to workflow data
- Implement custom compression for large payloads
- Support legacy data formats 