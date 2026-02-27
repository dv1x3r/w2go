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

type GetGridRequest struct {
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

func ParseGetGridRequest(request string) (GetGridRequest, error) {
	var req GetGridRequest
	return req, json.Unmarshal([]byte(request), &req)
}

type GetGridResponse[T any] struct {
	Status  Status `json:"status"`
	Records []T    `json:"records,omitempty"`
	Summary []T    `json:"summary,omitempty"`
	Total   int    `json:"total,omitempty"`
}

func NewGetGridResponse[T any](records []T, total int) GetGridResponse[T] {
	return GetGridResponse[T]{
		Status:  StatusSuccess,
		Records: records,
		Total:   total,
	}
}

func NewGetGridResponseWithSummary[T any](records []T, summary []T, total int) GetGridResponse[T] {
	return GetGridResponse[T]{
		Status:  StatusSuccess,
		Records: records,
		Summary: summary,
		Total:   total,
	}
}

func (res GetGridResponse[T]) Write(w http.ResponseWriter) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}

type SaveGridRequest[T any] struct {
	Changes []T `json:"changes"`
}

func ParseSaveGridRequest[T any](body io.Reader) (SaveGridRequest[T], error) {
	var req SaveGridRequest[T]
	return req, json.NewDecoder(body).Decode(&req)
}

type RemoveGridRequest struct {
	ID []int `json:"id"`
}

func ParseRemoveGridRequest(body io.Reader) (RemoveGridRequest, error) {
	var req RemoveGridRequest
	return req, json.NewDecoder(body).Decode(&req)
}

type GetFormRequest struct {
	Action string `json:"action"`
	Name   string `json:"name"`
	RecID  int    `json:"recid"`
}

func ParseGetFormRequest(request string) (GetFormRequest, error) {
	var req GetFormRequest
	return req, json.Unmarshal([]byte(request), &req)
}

type GetFormResponse[T any] struct {
	Status Status `json:"status"`
	Record *T     `json:"record,omitempty"`
}

func NewGetFormResponse[T any](record T) GetFormResponse[T] {
	return GetFormResponse[T]{
		Status: StatusSuccess,
		Record: &record,
	}
}

func (res GetFormResponse[T]) Write(w http.ResponseWriter) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}

type SaveFormRequest[T any] struct {
	Action string `json:"action"`
	Name   string `json:"name"`
	RecID  int    `json:"recid"`
	Record T      `json:"record"`
}

func ParseSaveFormRequest[T any](body io.Reader) (SaveFormRequest[T], error) {
	var req SaveFormRequest[T]
	return req, json.NewDecoder(body).Decode(&req)
}

type SaveFormResponse struct {
	Status Status `json:"status"`
	RecID  int    `json:"recid,omitempty"`
}

func NewSaveFormResponse(recID int) SaveFormResponse {
	return SaveFormResponse{
		Status: StatusSuccess,
		RecID:  recID,
	}
}

func (res SaveFormResponse) Write(w http.ResponseWriter) error {
	data, err := json.Marshal(res)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)
	return err
}
