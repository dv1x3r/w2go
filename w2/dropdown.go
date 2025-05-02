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

func (res DropdownResponse[T]) Write(w http.ResponseWriter) {
	data, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

type DropdownValue struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}
