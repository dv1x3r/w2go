package w2

import (
	"encoding/json"
	"io"
)

// ReorderGridRequest is the drag-and-drop reorder payload for moving one grid row.
type ReorderGridRequest struct {
	// RecID is the ID of the row being moved.
	RecID int

	// MoveBefore is the row ID that RecID should be moved before.
	MoveBefore int

	// Bottom is true when RecID should be moved to the end of the list.
	Bottom bool
}

// ParseReorderGridRequest decodes a w2grid reorder request body for one row.
func ParseReorderGridRequest(body io.Reader) (ReorderGridRequest, error) {
	var req ReorderGridRequest
	return req, json.NewDecoder(body).Decode(&req)
}

// MarshalJSON encodes the request using w2ui's "recid" and "moveBefore" field names.
func (req ReorderGridRequest) MarshalJSON() ([]byte, error) {
	v := struct {
		RecID      int `json:"recid"`
		MoveBefore any `json:"moveBefore"`
	}{}

	v.RecID = req.RecID

	if req.Bottom {
		v.MoveBefore = "bottom"
	} else {
		v.MoveBefore = req.MoveBefore
	}

	return json.Marshal(v)
}

// UnmarshalJSON decodes w2ui's "recid" and "moveBefore" fields.
//
// The moveBefore value can be either another row ID or the string "bottom".
func (req *ReorderGridRequest) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	raw := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if err := json.Unmarshal(raw["recid"], &req.RecID); err != nil {
		return err
	}

	if string(raw["moveBefore"]) == `"bottom"` {
		req.Bottom = true
	} else if err := json.Unmarshal(raw["moveBefore"], &req.MoveBefore); err != nil {
		return err
	}

	return nil
}

// ReorderGridArrayRequest is the drag-and-drop reorder payload for moving multiple grid rows together.
type ReorderGridArrayRequest struct {
	// RecID contains the IDs of the rows being moved.
	RecID []int

	// MoveBefore is the row ID that RecID should be moved before.
	MoveBefore int

	// Bottom is true when RecID should be moved to the end of the list.
	Bottom bool
}

// ParseReorderGridArrayRequest decodes a w2grid reorder request body for multiple rows.
func ParseReorderGridArrayRequest(body io.Reader) (ReorderGridArrayRequest, error) {
	var req ReorderGridArrayRequest
	return req, json.NewDecoder(body).Decode(&req)
}

// MarshalJSON encodes the request using w2ui's "recid" and "moveBefore" field names.
func (req ReorderGridArrayRequest) MarshalJSON() ([]byte, error) {
	v := struct {
		RecID      []int `json:"recid"`
		MoveBefore any   `json:"moveBefore"`
	}{}

	v.RecID = req.RecID

	if req.Bottom {
		v.MoveBefore = "bottom"
	} else {
		v.MoveBefore = req.MoveBefore
	}

	return json.Marshal(v)
}

// UnmarshalJSON decodes w2ui's multi-row reorder payload.
//
// The moveBefore value can be either another row ID or the string "bottom".
func (req *ReorderGridArrayRequest) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	raw := make(map[string]json.RawMessage)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if err := json.Unmarshal(raw["recid"], &req.RecID); err != nil {
		return err
	}

	if string(raw["moveBefore"]) == `"bottom"` {
		req.Bottom = true
	} else if err := json.Unmarshal(raw["moveBefore"], &req.MoveBefore); err != nil {
		return err
	}

	return nil
}
