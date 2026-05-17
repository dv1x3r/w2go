package w2

import (
	"database/sql"
	"database/sql/driver"
	"time"
)

// UnixTime stores a time.Time value as a Unix timestamp in SQL.
//
// A SQL NULL scans to the zero time, and the zero time writes back as SQL NULL.
type UnixTime struct {
	time.Time
}

// Scan implements sql.Scanner for integer Unix timestamps in seconds.
func (t *UnixTime) Scan(value any) error {
	var n sql.NullInt64
	if err := n.Scan(value); err != nil {
		return err
	}

	if n.Valid {
		t.Time = time.Unix(n.Int64, 0).UTC()
	} else {
		t.Time = time.Time{}
	}

	return nil
}

// Value implements driver.Valuer and returns a Unix timestamp in seconds.
func (t UnixTime) Value() (driver.Value, error) {
	if t.Time.IsZero() {
		return nil, nil
	}
	return t.UTC().Unix(), nil
}
