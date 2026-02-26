package w2

import (
	"encoding/json"
	"io"
)

type ReorderGridRequest struct {
	RecID      int
	MoveBefore int
	Bottom     bool
}

func ParseReorderGridRequest(body io.Reader) (ReorderGridRequest, error) {
	var req ReorderGridRequest
	return req, json.NewDecoder(body).Decode(&req)
}

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

type ReorderManyGridRequest struct {
	RecID      []int
	MoveBefore int
	Bottom     bool
}

func ParseReorderManyGridRequest(body io.Reader) (ReorderManyGridRequest, error) {
	var req ReorderManyGridRequest
	return req, json.NewDecoder(body).Decode(&req)
}

func (req ReorderManyGridRequest) MarshalJSON() ([]byte, error) {
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

func (req *ReorderManyGridRequest) UnmarshalJSON(data []byte) error {
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
