package ibcmsg

import (
	"encoding/base64"
	"reflect"
	"time"

	"github.com/crypto-com/chain-indexing/usecase/model"

	"github.com/mitchellh/mapstructure"
)

// []byte is JSON encoded as base64 string in Golang. This decode function converts string -> []byte pair correctly.
func StringToByteSliceHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.SliceOf(reflect.TypeOf(byte('1'))) {
			return data, nil
		}

		// []byte is encoded as base64 string
		decoded, decodeErr := base64.StdEncoding.DecodeString(data.(string))
		if decodeErr != nil {
			return nil, decodeErr
		}
		return decoded, nil
	}
}

// String to wrapped Duration
func StringToDurationHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != reflect.TypeOf(model.Duration{}) {
			return data, nil
		}

		// Convert it by parsing
		d, parseErr := time.ParseDuration(data.(string))
		if parseErr != nil {
			return nil, parseErr
		}
		return model.Duration{Duration: d}, nil
	}
}
