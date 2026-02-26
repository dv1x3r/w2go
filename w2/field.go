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
func (f Field[T]) IsZero() bool {
	return !f.Provided
}

// UnmarshalJSON implements the json.Unmarshaler interface.
// Ensures proper unmarshalling of JSON data, setting the Valid flag based on the JSON data.
func (f *Field[T]) UnmarshalJSON(data []byte) error {
	f.Provided = true

	// w2grid inline edit sends empty string for blank fields
	if string(data) == "null" || string(data) == `""` {
		f.Valid = false
		return nil
	}

	err := json.Unmarshal(data, &f.V)
	f.Valid = err == nil
	return err
}

// MarshalJSON implements the json.Marshaler interface.
// Ensures proper marshalling of Field values, representing unset values as null.
func (f Field[T]) MarshalJSON() ([]byte, error) {
	if !f.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(f.V)
}

// Scan implements the sql.Scanner interface for Field.
// Ensures that Provided flag is set.
func (f *Field[T]) Scan(value any) error {
	f.Provided = true
	return f.Null.Scan(value)
}

// Value implements the driver.Valuer interface for Field.
func (f Field[T]) Value() (driver.Value, error) {
	return f.Null.Value()
}
