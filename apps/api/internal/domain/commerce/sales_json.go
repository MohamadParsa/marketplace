package commerce

import (
	"bytes"
	"encoding/json"
	"errors"
)

// Missing/null is not an explicit zero override. Deleting the override row
// restores inheritance; a valid row always includes its numeric rate.
func (v *CategoryMarkup) UnmarshalJSON(data []byte) error {
	var in struct {
		CategoryID string `json:"category_id"`
		MethodID   string `json:"method_id"`
		MarkupBPS  *int   `json:"markup_bps"`
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if e := d.Decode(&in); e != nil {
		return e
	}
	if in.MarkupBPS == nil {
		return errors.New("explicit markup rate required")
	}
	*v = CategoryMarkup{CategoryID: in.CategoryID, MethodID: in.MethodID, MarkupBPS: *in.MarkupBPS}
	return nil
}
