package w2

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

// Field wraps sql.Null[T] and tracks whether the value was explicitly provided (e.g., via JSON or SQL).
type Field[T any] struct {
	sql.Null[T]
	Provided bool
}

// NewField creates a new Field[T] with a given value, marked as valid and provided.
func NewField[T any](value T) Field[T] {
	return Field[T]{
		Null: sql.Null[T]{
			V:     value,
			Valid: true,
		},
		Provided: true,
	}
}

// IsZero implements the json.Zeroed interface.
// Ensures proper marshalling of Field values, skipping non-provided values with omitzero tag.
func (e Field[T]) IsZero() bool {
	return !e.Provided
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// Ensures proper unmarshalling of JSON data, setting the Valid flag based on the JSON data.
func (e *Field[T]) UnmarshalJSON(data []byte) error {
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
// Ensures proper marshalling of Field values, representing unset values as null.
func (e Field[T]) MarshalJSON() ([]byte, error) {
	if !e.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(e.V)
}

// Scan implements the sql.Scanner interface for Field.
// Ensures that Provided flag is set.
func (e *Field[T]) Scan(value any) error {
	e.Provided = true
	return e.Null.Scan(value)
}

// Value implements the driver.Valuer interface for Field.
func (e Field[T]) Value() (driver.Value, error) {
	return e.Null.Value()
}
