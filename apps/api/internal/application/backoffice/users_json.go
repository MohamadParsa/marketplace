package backoffice

import (
	"bytes"
	"encoding/json"
)

func (in *UserInput) UnmarshalJSON(data []byte) error {
	type plain UserInput
	var value plain
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if e := d.Decode(&value); e != nil {
		return e
	}
	var fields map[string]json.RawMessage
	if e := json.Unmarshal(data, &fields); e != nil {
		return e
	}
	for _, key := range []string{"active", "version"} {
		v, ok := fields[key]
		if !ok || bytes.Equal(bytes.TrimSpace(v), []byte("null")) {
			return ErrInvalid
		}
	}
	*in = UserInput(value)
	return nil
}
