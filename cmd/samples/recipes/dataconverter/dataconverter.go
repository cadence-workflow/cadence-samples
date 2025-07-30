package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"go.uber.org/cadence/encoded"
)

type compressedJSONDataConverter struct{}

func NewCompressedJSONDataConverter() encoded.DataConverter {
	return &compressedJSONDataConverter{}
}

func (dc *compressedJSONDataConverter) ToData(value ...interface{}) ([]byte, error) {
	// First, serialize to JSON
	var jsonBuf bytes.Buffer
	enc := json.NewEncoder(&jsonBuf)
	for i, obj := range value {
		err := enc.Encode(obj)
		if err != nil {
			return nil, fmt.Errorf("unable to encode argument: %d, %v, with error: %v", i, reflect.TypeOf(obj), err)
		}
	}

	// Then compress the JSON data
	var compressedBuf bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedBuf)

	_, err := gzipWriter.Write(jsonBuf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("unable to compress data: %v", err)
	}

	err = gzipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to close gzip writer: %v", err)
	}

	return compressedBuf.Bytes(), nil
}

func (dc *compressedJSONDataConverter) FromData(input []byte, valuePtr ...interface{}) error {
	// First, decompress the data
	gzipReader, err := gzip.NewReader(bytes.NewBuffer(input))
	if err != nil {
		return fmt.Errorf("unable to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	// Read the decompressed JSON data
	decompressedData, err := io.ReadAll(gzipReader)
	if err != nil {
		return fmt.Errorf("unable to decompress data: %v", err)
	}

	// Then deserialize from JSON
	dec := json.NewDecoder(bytes.NewBuffer(decompressedData))
	for i, obj := range valuePtr {
		err := dec.Decode(obj)
		if err != nil {
			return fmt.Errorf("unable to decode argument: %d, %v, with error: %v", i, reflect.TypeOf(obj), err)
		}
	}
	return nil
}
