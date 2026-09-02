package marshaler

import (
	"bytes"
	"io"

	"github.com/goccy/go-json"
)

// LoadFromReader load data from raw byte reader to given interface.
func LoadJSONFromReader(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

// MarshalToString Marshal to string interface.
func MarshalJSONToString(v any) (string, error) {
	res, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

// Marshal Marshal to byte array.
func MarshalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

// Marshal Marshal to string without escape.
func MarshalJSONToStringWithoutEscape(v any) (string, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)

	return buffer.String(), err
}

// Marshal Marshal to byte array without escape.
func MarshalJSONWithoutEscape(v any) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(v)

	return buffer.Bytes(), err
}

// UnmarshalFromString unmarshal from string.
func UnmarshalJSONFromString(str string, v any) error {
	return json.Unmarshal([]byte(str), v)
}

// Unmarsha unmarshal from byte array.
func UnmarshalJSON(b []byte, v any) error {
	return json.Unmarshal(b, v)
}
