package w2

import (
	"encoding/json"
	"io"
	"net/http"
)

type Status string

const (
	StatusSuccess Status = "success"
	StatusError   Status = "error"
)

type BaseResponse struct {
	Status  Status `json:"status"`
	Message string `json:"message,omitempty"`
}

func NewSuccessResponse() BaseResponse {
	return BaseResponse{Status: StatusSuccess}
}

func NewErrorResponse(message string) BaseResponse {
	return BaseResponse{Status: StatusError, Message: message}
}

func (res BaseResponse) Write(w http.ResponseWriter, statusCode int) {
	data, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(data)
}

type GridDataRequest struct {
	Limit       int          `json:"limit"`
	Offset      int          `json:"offset"`
	SearchLogic string       `json:"searchLogic"`
	Search      []GridSearch `json:"search"`
	Sort        []GridSort   `json:"sort"`
}

type GridSearch struct {
	Field    string `json:"field"`
	Type     string `json:"type"`
	Operator string `json:"operator"`
	Value    any    `json:"value"`
}

type GridSort struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

func ParseGridDataRequest(request string) (GridDataRequest, error) {
	var req GridDataRequest
	return req, json.Unmarshal([]byte(request), &req)
}

type GridDataResponse[T any, V any] struct {
	Status  Status `json:"status"`
	Records []T    `json:"records,omitempty"`
	Summary []V    `json:"summary,omitempty"`
	Total   int    `json:"total,omitempty"`
}

func NewGridDataResponse[T any](records []T, total int) GridDataResponse[T, any] {
	return GridDataResponse[T, any]{
		Status:  StatusSuccess,
		Records: records,
		Total:   total,
	}
}

func NewGridDataResponseWithSummary[T any, V any](records []T, summary []V, total int) GridDataResponse[T, V] {
	return GridDataResponse[T, V]{
		Status:  StatusSuccess,
		Records: records,
		Summary: summary,
		Total:   total,
	}
}

func (res GridDataResponse[T, V]) Write(w http.ResponseWriter) {
	data, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

type GridSaveRequest[T any] struct {
	Changes []T `json:"changes"`
}

func ParseGridSaveRequest[T any](body io.Reader) (GridSaveRequest[T], error) {
	var req GridSaveRequest[T]
	return req, json.NewDecoder(body).Decode(&req)
}

type GridRemoveRequest struct {
	ID []int `json:"id"`
}

func ParseGridRemoveRequest(body io.Reader) (GridRemoveRequest, error) {
	var req GridRemoveRequest
	return req, json.NewDecoder(body).Decode(&req)
}

type FormGetRequest struct {
	Action string `json:"action"`
	Name   string `json:"name"`
	RecID  int    `json:"recid"`
}

func ParseFormGetRequest(request string) (FormGetRequest, error) {
	var req FormGetRequest
	return req, json.Unmarshal([]byte(request), &req)
}

type FormGetResponse[T any] struct {
	Status Status `json:"status"`
	Record *T     `json:"record,omitempty"`
}

func NewFormGetResponse[T any](record T) FormGetResponse[T] {
	return FormGetResponse[T]{
		Status: StatusSuccess,
		Record: &record,
	}
}

func (res FormGetResponse[T]) Write(w http.ResponseWriter) {
	data, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

type FormSaveRequest[T any] struct {
	Action string `json:"action"`
	Name   string `json:"name"`
	RecID  int    `json:"recid"`
	Record T      `json:"record"`
}

func ParseFormSaveRequest[T any](body io.Reader) (FormSaveRequest[T], error) {
	var req FormSaveRequest[T]
	return req, json.NewDecoder(body).Decode(&req)
}

type FormSaveResponse struct {
	Status Status `json:"status"`
	RecID  int    `json:"recid,omitempty"`
}

func NewFormSaveResponse(recID int) FormSaveResponse {
	return FormSaveResponse{
		Status: StatusSuccess,
		RecID:  recID,
	}
}

func (res FormSaveResponse) Write(w http.ResponseWriter) {
	data, _ := json.Marshal(res)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
