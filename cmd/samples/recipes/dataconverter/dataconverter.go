package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"go.uber.org/cadence/encoded"
)

type jsonDataConverter struct{}

func NewJSONDataConverter() encoded.DataConverter {
	return &jsonDataConverter{}
}

func (dc *jsonDataConverter) ToData(value ...interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	for i, obj := range value {
		err := enc.Encode(obj)
		if err != nil {
			return nil, fmt.Errorf("unable to encode argument: %d, %v, with error: %v", i, reflect.TypeOf(obj), err)
		}
	}
	return buf.Bytes(), nil
}

func (dc *jsonDataConverter) FromData(input []byte, valuePtr ...interface{}) error {
	dec := json.NewDecoder(bytes.NewBuffer(input))
	for i, obj := range valuePtr {
		err := dec.Decode(obj)
		if err != nil {
			return fmt.Errorf("unable to decode argument: %d, %v, with error: %v", i, reflect.TypeOf(obj), err)
		}
	}
	return nil
}
