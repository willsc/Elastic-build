package object

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRawEncoders(t *testing.T) {
	body, err := json.Marshal(map[string]interface{}{
		"hello": "world",
		"field": 1.23,
	})
	require.NoError(t, err)

	raw := &RawObject{
		UID:     "a",
		Kind:    "b",
		Version: "c",
		ETag:    "d",
		Body:    body,
	}

	b, err := json.MarshalIndent(raw, "", "  ")
	require.NoError(t, err)

	str := string(b)
	require.JSONEq(t, `{"UID":"a","kind":"b","version":"c","body":{"field":1.23,"hello":"world"},"etag":"d"}`, str)

	copy := &RawObject{}
	err = json.Unmarshal(b, copy)
	require.NoError(t, err)
}
