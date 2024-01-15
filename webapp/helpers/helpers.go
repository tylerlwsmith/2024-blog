package helpers

import (
	"encoding/json"
	"io"
)

func UnmarshalJsonReader[T any](r io.Reader, v *T) (err error) {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &v)
	if err != nil {
		return err
	}

	return nil
}
