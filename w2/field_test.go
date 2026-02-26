package w2_test

import (
	"encoding/json"
	"testing"

	"github.com/dv1x3r/w2go/w2"
)

type Todo struct {
	ID       int              `json:"id"`
	Name     w2.Field[string] `json:"name,omitzero"`
	Quantity w2.Field[int]    `json:"quantity,omitzero"`
}

func TestField(t *testing.T) {
	t.Run("JSONRoundTrip", func(t *testing.T) {
		tests := []struct {
			InputJSON    string
			Expected     Todo
			ExpectedJSON string
		}{
			{
				InputJSON: `{"id": 1, "name": "Buy milk", "quantity": 2}`,
				Expected: Todo{
					ID:       1,
					Name:     w2.NewField("Buy milk"),
					Quantity: w2.NewField(2),
				},
				ExpectedJSON: `{"id":1,"name":"Buy milk","quantity":2}`,
			},
			{
				InputJSON: `{"id": 2, "name": null}`,
				Expected: Todo{
					ID:       2,
					Name:     w2.Field[string]{Provided: true},
					Quantity: w2.Field[int]{Provided: false},
				},
				ExpectedJSON: `{"id":2,"name":null}`,
			},
			{
				InputJSON: `{"id": 3}`,
				Expected: Todo{
					ID:       3,
					Name:     w2.Field[string]{},
					Quantity: w2.Field[int]{},
				},
				ExpectedJSON: `{"id":3}`,
			},
		}

		for _, test := range tests {
			var value Todo
			err := json.Unmarshal([]byte(test.InputJSON), &value)
			if err != nil {
				t.Errorf("❌ Unmarshal error for input %s: %v", test.InputJSON, err)
				continue
			}

			if value != test.Expected {
				t.Errorf("❌ Unexpected struct for input %s:\n  got:  %+v\n  want: %+v", test.InputJSON, value, test.Expected)
				continue
			}

			output, err := json.Marshal(value)
			if err != nil {
				t.Errorf("❌ Marshal error for struct %+v: %v", value, err)
				continue
			}

			if string(output) != test.ExpectedJSON {
				t.Errorf("❌ Unexpected output JSON:\n  got:  %s\n  want: %s", output, test.ExpectedJSON)
			}
		}
	})
}
