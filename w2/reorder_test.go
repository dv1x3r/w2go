package w2_test

import (
	"encoding/json"
	"testing"

	"github.com/dv1x3r/w2go/w2"
)

func TestGridReorderRequest(t *testing.T) {
	t.Run("JSONRoundTrip", func(t *testing.T) {
		tests := []struct {
			InputJSON    string
			Expected     w2.GridReorderRequest
			ExpectedJSON string
		}{
			{
				InputJSON:    `{"recid": 1, "moveBefore": 42}`,
				Expected:     w2.GridReorderRequest{RecID: 1, MoveBefore: 42, Last: false},
				ExpectedJSON: `{"recid":1,"moveBefore":42}`,
			},
			{
				InputJSON:    `{"recid": 2, "moveBefore": "bottom"}`,
				Expected:     w2.GridReorderRequest{RecID: 2, Last: true},
				ExpectedJSON: `{"recid":2,"moveBefore":"bottom"}`,
			},
		}

		for _, test := range tests {
			var req w2.GridReorderRequest
			err := json.Unmarshal([]byte(test.InputJSON), &req)
			if err != nil {
				t.Errorf("❌ Unmarshal error for input %s: %v", test.InputJSON, err)
				continue
			}

			if req != test.Expected {
				t.Errorf("❌ Unexpected struct for input %s:\n  got:  %+v\n  want: %+v", test.InputJSON, req, test.Expected)
			}

			output, err := json.Marshal(req)
			if err != nil {
				t.Errorf("❌ Marshal error for input %s: %v", test.InputJSON, err)
				continue
			}

			if string(output) != test.ExpectedJSON {
				t.Errorf("❌ Unexpected output JSON:\n  got:  %s\n  want: %s", output, test.ExpectedJSON)
			}
		}
	})
}
