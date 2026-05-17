package w2

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// Field wraps sql.Null[T] and tracks whether the value was explicitly provided.
//
// It is designed for w2grid inline edits, where unchanged fields are omitted
// from JSON. Provided reports whether the client sent the field. Valid reports
// whether the sent value should be stored as a non-NULL value.
type Field[T any] struct {
	sql.Null[T]

	// Provided is true when the field was present in JSON or read from SQL.
	Provided bool
}

// NewField returns a Field containing value and marks it as both provided and valid.
func NewField[T any](value T) Field[T] {
	return Field[T]{
		Null: sql.Null[T]{
			V:     value,
			Valid: true,
		},
		Provided: true,
	}
}

// NotNull returns a copy that writes the zero value instead of SQL NULL.
//
// Use it when an empty value from w2ui should be saved as the type's zero value rather than NULL.
func (f Field[T]) NotNull() Field[T] {
	f.Valid = true
	return f
}

// IsProvided implements the w2db.Providable interface.
//
// It lets SQL helpers skip fields that were not sent by the client.
func (f Field[T]) IsProvided() bool {
	return f.Provided
}

// IsZero reports whether the field was not provided.
//
// It lets encoding/json omit non-provided fields when the struct tag uses ",omitzero".
func (f Field[T]) IsZero() bool {
	return !f.Provided
}

// UnmarshalJSON implements json.Unmarshaler and marks the field as provided.
//
// JSON null and an empty string both produce a provided but invalid field,
// matching how w2grid sends blank inline-edit values.
func (f *Field[T]) UnmarshalJSON(data []byte) error {
	f.Provided = true

	// w2grid inline edit sends empty string for blank fields
	if string(data) == "null" || string(data) == `""` {
		var zero T
		f.V = zero
		f.Valid = false
		return nil
	}

	err := json.Unmarshal(data, &f.V)
	f.Valid = err == nil
	return err
}

// MarshalJSON implements json.Marshaler.
//
// Invalid fields are encoded as JSON null. To omit a field that was not
// provided, use the ",omitzero" struct tag.
func (f Field[T]) MarshalJSON() ([]byte, error) {
	if !f.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(f.V)
}

// Scan implements sql.Scanner and marks the field as provided.
func (f *Field[T]) Scan(value any) error {
	f.Provided = true
	return f.Null.Scan(value)
}

// Value implements driver.Valuer for database writes.
func (f Field[T]) Value() (driver.Value, error) {
	return f.Null.Value()
}

// String implements fmt.Stringer for debugging and logs.
func (f Field[T]) String() string {
	if !f.Valid {
		return "<nil>"
	}
	return fmt.Sprint(f.V)
}
