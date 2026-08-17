package w2

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

var unixTimeFallbackJSONLayouts = []struct {
	layout   string
	location *time.Location
}{
	{"2006-01-02 15:04:05Z07:00", nil},
	{"2006-01-02T15:04:05", time.UTC},
	{"2006-01-02T15:04", time.UTC},
	{"2006-01-02 15:04:05", time.UTC},
	{"2006-01-02 15:04", time.UTC},
	{"2006-01-02", time.UTC},
}

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

// UnmarshalJSON implements json.Unmarshaler for JSON date/time strings.
func (t *UnixTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		t.Time = time.Time{}
		return nil
	}

	var parsed time.Time
	if err := parsed.UnmarshalJSON(data); err == nil {
		t.Time = parsed.UTC()
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	for _, candidate := range unixTimeFallbackJSONLayouts {
		var parsed time.Time
		var err error

		if candidate.location == nil {
			parsed, err = time.Parse(candidate.layout, value)
		} else {
			parsed, err = time.ParseInLocation(candidate.layout, value, candidate.location)
		}
		if err != nil {
			continue
		}

		t.Time = parsed.UTC()
		return nil
	}

	return fmt.Errorf("w2.UnixTime: cannot parse %q as a JSON date or date-time", value)
}
