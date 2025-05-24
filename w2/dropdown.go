package w2

import (
	"encoding/json"
	"net/http"
)

type DropdownRequest struct {
	Max    int    `json:"max"`
	Search string `json:"search"`
}

func ParseDropdownRequest(request string) (DropdownRequest, error) {
	var req DropdownRequest
	return req, json.Unmarshal([]byte(request), &req)
}

type DropdownResponse[T any] struct {
	Status  Status `json:"status"`
	Records []T    `json:"records"`
}

func NewDropdownResponse[T any](records []T) DropdownResponse[T] {
	return DropdownResponse[T]{
		Status:  StatusSuccess,
		Records: records,
	}
}

func (res DropdownResponse[T]) Write(w http.ResponseWriter) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}

type DropdownValue struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type EditableDropdown struct {
	ID   Editable[int]    `json:"id"`
	Text Editable[string] `json:"text"`
}

func (e *EditableDropdown) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		e.ID.Provided = true
		e.Text.Provided = true
		return nil
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if err := json.Unmarshal(raw["id"], &e.ID); err != nil {
		return err
	}

	if err := json.Unmarshal(raw["text"], &e.Text); err != nil {
		return err
	}

	return nil
}
