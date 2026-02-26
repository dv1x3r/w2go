package w2

import (
	"encoding/json"
	"io"
)

type GridReorderRequest struct {
	RecID      int
	MoveBefore int
	Bottom     bool
}

func ParseGridReorderRequest(body io.Reader) (GridReorderRequest, error) {
	var req GridReorderRequest
	return req, json.NewDecoder(body).Decode(&req)
}

func (req GridReorderRequest) MarshalJSON() ([]byte, error) {
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

func (req *GridReorderRequest) UnmarshalJSON(data []byte) error {
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

type GridReorderManyRequest struct {
	RecID      []int
	MoveBefore int
	Bottom     bool
}

func ParseGridReorderManyRequest(body io.Reader) (GridReorderManyRequest, error) {
	var req GridReorderManyRequest
	return req, json.NewDecoder(body).Decode(&req)
}

func (req GridReorderManyRequest) MarshalJSON() ([]byte, error) {
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

func (req *GridReorderManyRequest) UnmarshalJSON(data []byte) error {
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
