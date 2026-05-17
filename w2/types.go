package w2

import (
	"encoding/json"
	"io"
	"net/http"
)

// Status is the w2ui response status string.
type Status string

const (
	// StatusSuccess marks a response as successful for w2ui.
	StatusSuccess Status = "success"

	// StatusError marks a response as failed and should usually include a message.
	StatusError Status = "error"
)

// BaseResponse is the common success/error response used when a widget action does not need to return records.
type BaseResponse struct {
	// Status is usually StatusSuccess or StatusError.
	Status Status `json:"status"`

	// Message is shown by callers when Status is StatusError.
	Message string `json:"message,omitempty"`
}

// NewSuccessResponse returns a basic successful w2ui response.
func NewSuccessResponse() BaseResponse {
	return BaseResponse{Status: StatusSuccess}
}

// NewErrorResponse returns a basic failed w2ui response with a user-facing message.
func NewErrorResponse(message string) BaseResponse {
	return BaseResponse{Status: StatusError, Message: message}
}

// Write sends the response as application/json using statusCode.
func (res BaseResponse) Write(w http.ResponseWriter, statusCode int) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err = w.Write(data)
	return err
}

// GetGridRequest is the request payload w2grid sends when it loads records.
type GetGridRequest struct {
	// Limit is the requested page size. A zero value means no limit was sent.
	Limit int `json:"limit"`

	// Offset is the zero-based row offset requested by the grid.
	Offset int `json:"offset"`

	// SearchLogic is usually "AND" or "OR" and controls how Search terms combine.
	SearchLogic string `json:"searchLogic"`

	// Search contains the grid filter rules.
	Search []GridSearch `json:"search"`

	// Sort contains the requested sort columns.
	Sort []GridSort `json:"sort"`
}

// GridSearch describes one w2grid search/filter rule.
type GridSearch struct {
	// Field is the client-side field name.
	Field string `json:"field"`

	// Type is the w2ui field type.
	Type string `json:"type"`

	// Operator is the w2ui search operator.
	Operator string `json:"operator"`

	// Value is the raw JSON-decoded search value for Operator.
	Value any `json:"value"`
}

// GridSort describes one w2grid sort rule.
type GridSort struct {
	// Field is the client-side field name.
	Field string `json:"field"`

	// Direction is "asc" or "desc".
	Direction string `json:"direction"`
}

// ParseGetGridRequest decodes the JSON value from w2grid's "request" query parameter.
func ParseGetGridRequest(request string) (GetGridRequest, error) {
	var req GetGridRequest
	return req, json.Unmarshal([]byte(request), &req)
}

// GetGridResponse is the JSON response expected by w2grid when loading records.
type GetGridResponse[T any] struct {
	// Status is set to StatusSuccess by the constructor helpers.
	Status Status `json:"status"`

	// Records are the current page of grid rows.
	Records []T `json:"records,omitempty"`

	// Summary contains optional footer/summary rows.
	Summary []T `json:"summary,omitempty"`

	// Total is the total number of rows after filtering, before pagination.
	Total int `json:"total,omitempty"`
}

// NewGetGridResponse returns a successful grid response with records and total.
func NewGetGridResponse[T any](records []T, total int) GetGridResponse[T] {
	return GetGridResponse[T]{
		Status:  StatusSuccess,
		Records: records,
		Total:   total,
	}
}

// NewGetGridResponseWithSummary returns a successful grid response with records, summary rows, and total.
func NewGetGridResponseWithSummary[T any](records []T, summary []T, total int) GetGridResponse[T] {
	return GetGridResponse[T]{
		Status:  StatusSuccess,
		Records: records,
		Summary: summary,
		Total:   total,
	}
}

// Write sends the grid response as application/json.
func (res GetGridResponse[T]) Write(w http.ResponseWriter) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}

// SaveGridRequest is the request body w2grid sends when saving inline edits.
type SaveGridRequest[T any] struct {
	// Changes contains the edited records. Use Field values on T to detect omitted columns.
	Changes []T `json:"changes"`
}

// ParseSaveGridRequest decodes a w2grid save request body.
func ParseSaveGridRequest[T any](body io.Reader) (SaveGridRequest[T], error) {
	var req SaveGridRequest[T]
	return req, json.NewDecoder(body).Decode(&req)
}

// RemoveGridRequest is the request body w2grid sends when deleting records.
type RemoveGridRequest struct {
	// ID contains the record IDs selected for deletion.
	ID []int `json:"id"`
}

// ParseRemoveGridRequest decodes a w2grid delete request body.
func ParseRemoveGridRequest(body io.Reader) (RemoveGridRequest, error) {
	var req RemoveGridRequest
	return req, json.NewDecoder(body).Decode(&req)
}

// GetFormRequest is the request payload w2form sends when loading a record.
type GetFormRequest struct {
	// Action is the w2ui action name.
	Action string `json:"action"`

	// Name is the w2form name.
	Name string `json:"name"`

	// RecID is the record ID requested by the form.
	RecID int `json:"recid"`
}

// ParseGetFormRequest decodes the JSON value from w2form's "request" query parameter.
func ParseGetFormRequest(request string) (GetFormRequest, error) {
	var req GetFormRequest
	return req, json.Unmarshal([]byte(request), &req)
}

// GetFormResponse is the JSON response expected by w2form when loading a single record.
type GetFormResponse[T any] struct {
	// Status is set to StatusSuccess by NewGetFormResponse.
	Status Status `json:"status"`

	// Record is the form record.
	Record *T `json:"record,omitempty"`
}

// NewGetFormResponse returns a successful form response for record.
func NewGetFormResponse[T any](record T) GetFormResponse[T] {
	return GetFormResponse[T]{
		Status: StatusSuccess,
		Record: &record,
	}
}

// Write sends the form response as application/json.
func (res GetFormResponse[T]) Write(w http.ResponseWriter) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}

// SaveFormRequest is the request body w2form sends when saving a record.
type SaveFormRequest[T any] struct {
	// Action is the w2ui action name.
	Action string `json:"action"`

	// Name is the w2form name.
	Name string `json:"name"`

	// RecID is the record ID being saved.
	RecID int `json:"recid"`

	// Record is the submitted form data.
	Record T `json:"record"`
}

// ParseSaveFormRequest decodes a w2form save request body.
func ParseSaveFormRequest[T any](body io.Reader) (SaveFormRequest[T], error) {
	var req SaveFormRequest[T]
	return req, json.NewDecoder(body).Decode(&req)
}

// SaveFormResponse is the JSON response expected by w2form after saving.
type SaveFormResponse struct {
	// Status is set to StatusSuccess by NewSaveFormResponse.
	Status Status `json:"status"`

	// RecID is the saved record ID. For inserts, set it to the new ID.
	RecID int `json:"recid,omitempty"`
}

// NewSaveFormResponse returns a successful form-save response.
func NewSaveFormResponse(recID int) SaveFormResponse {
	return SaveFormResponse{
		Status: StatusSuccess,
		RecID:  recID,
	}
}

// Write sends the form-save response as application/json.
func (res SaveFormResponse) Write(w http.ResponseWriter) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}
