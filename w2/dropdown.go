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

type Dropdown struct {
	ID   Field[int]    `json:"id"`
	Text Field[string] `json:"text"`
}

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
