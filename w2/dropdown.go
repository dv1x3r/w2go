package w2

import (
	"encoding/json"
	"net/http"
)

// Dropdown is the standard record shape for w2ui dropdown/list options.
type Dropdown struct {
	// ID is the option value.
	ID Field[int] `json:"id"`

	// Text is the option label shown to the user.
	Text Field[string] `json:"text"`
}

// UnmarshalJSON accepts the common w2ui dropdown encodings.
//
// w2ui may submit a selected item as a bare integer ID, as an object containing
// id and text, as null, or as an empty string. The ID and Text fields use Field
// so callers can tell whether a value was provided.
func (d *Dropdown) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		d.ID.Provided = true
		d.Text.Provided = true
		return nil
	}

	// parse integer with ID
	// - w2form saveCleanRecord is true (default)
	if err := json.Unmarshal(data, &d.ID); err == nil {
		return nil
	}

	// parse object with ID and Text
	// - w2form saveCleanRecord is false
	// - w2grid editable dropdown list
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if err := json.Unmarshal(raw["id"], &d.ID); err != nil {
		return err
	}

	if err := json.Unmarshal(raw["text"], &d.Text); err != nil {
		return err
	}

	return nil
}

// GetDropdownRequest is the request payload w2ui sends when loading dropdown
// options.
type GetDropdownRequest struct {
	// Max is the maximum number of options requested.
	Max int `json:"max"`

	// Search is the user's search text.
	Search string `json:"search"`
}

// ParseGetDropdownRequest decodes the JSON value from a dropdown "request"
// query parameter.
func ParseGetDropdownRequest(request string) (GetDropdownRequest, error) {
	var req GetDropdownRequest
	return req, json.Unmarshal([]byte(request), &req)
}

// GetDropdownResponse is the JSON response expected by w2ui dropdown controls.
type GetDropdownResponse[T any] struct {
	// Status is set to StatusSuccess by NewGetDropdownResponse.
	Status Status `json:"status"`

	// Records are the dropdown options.
	Records []T `json:"records"`
}

// NewGetDropdownResponse returns a successful dropdown response.
func NewGetDropdownResponse[T any](records []T) GetDropdownResponse[T] {
	return GetDropdownResponse[T]{
		Status:  StatusSuccess,
		Records: records,
	}
}

// Write sends the dropdown response as application/json.
func (res GetDropdownResponse[T]) Write(w http.ResponseWriter) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}
