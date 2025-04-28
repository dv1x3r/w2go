package w2

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

// Editable wraps sql.Null[T] and tracks whether the value was explicitly provided (e.g., via JSON or SQL).
type Editable[T any] struct {
	sql.Null[T]
	Provided bool
}

// NewEditable creates a new Editable[T] with Provided set to true but no value set.
func NewEditable[T any]() Editable[T] {
	return Editable[T]{Provided: true}
}

// NewEditableWithValue creates a new Editable[T] with a given value, marked as valid and provided.
func NewEditableWithValue[T any](value T) Editable[T] {
	return Editable[T]{
		Null: sql.Null[T]{
			V:     value,
			Valid: true,
		},
		Provided: true,
	}
}

// IsZero implements the json.Zeroed interface.
// Ensures proper marshalling of Editable values, skipping non-provided values with omitzero tag.
func (e Editable[T]) IsZero() bool {
	return !e.Provided
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// Ensures proper unmarshalling of JSON data, setting the Valid flag based on the JSON data.
func (e *Editable[T]) UnmarshalJSON(data []byte) error {
	e.Provided = true

	// w2grid inline edit sends empty string for blank fields
	if string(data) == "null" || string(data) == `""` {
		e.Valid = false
		return nil
	}

	err := json.Unmarshal(data, &e.V)
	e.Valid = err == nil
	return err
}

// MarshalJSON implements the json.Marshaler interface.
// Ensures proper marshalling of Editable values, representing unset values as null.
func (e Editable[T]) MarshalJSON() ([]byte, error) {
	if !e.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(e.V)
}

// Scan implements the sql.Scanner interface for Editable.
// Ensures that Provided flag is set.
func (e *Editable[T]) Scan(value any) error {
	e.Provided = true
	return e.Null.Scan(value)
}

// Value implements the driver.Valuer interface for Editable.
func (e Editable[T]) Value() (driver.Value, error) {
	return e.Null.Value()
}
