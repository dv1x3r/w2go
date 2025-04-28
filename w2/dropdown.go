package w2

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
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

// Scan implements the sql.Scanner interface.
func (v *DropdownValue) Scan(src any) error {
	if src == nil {
		v.ID = 0
		return nil
	}

	if value, ok := src.(int64); ok {
		v.ID = int(value)
		return nil
	}

	return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type int64", src)
}

// Value implements the driver.Valuer interface.
func (v DropdownValue) Value() (driver.Value, error) {
	return int64(v.ID), nil
}
