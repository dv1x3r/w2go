package w2sort_test

import (
	"slices"
	"testing"

	"github.com/dv1x3r/w2go/w2"
	"github.com/dv1x3r/w2go/w2sort"
)

func TestReorderArray(t *testing.T) {
	t.Run("ReorderArray", func(t *testing.T) {
		tests := []struct {
			Input         []int
			Request       w2.GridReorderRequest
			Expected      []int
			ExpectedError bool
		}{
			{
				Input:    []int{1, 2, 3, 4, 5},
				Request:  w2.GridReorderRequest{RecID: 2, MoveBefore: 4},
				Expected: []int{1, 3, 2, 4, 5},
			},
			{
				Input:    []int{1, 2, 3, 4, 5},
				Request:  w2.GridReorderRequest{RecID: 4, MoveBefore: 1},
				Expected: []int{4, 1, 2, 3, 5},
			},
			{
				Input:    []int{1, 2, 3, 4, 5},
				Request:  w2.GridReorderRequest{RecID: 3, Bottom: true},
				Expected: []int{1, 2, 4, 5, 3},
			},
			{
				Input:         []int{1, 2, 3},
				Request:       w2.GridReorderRequest{RecID: 9, MoveBefore: 1},
				ExpectedError: true,
			},
			{
				Input:         []int{1, 2, 3},
				Request:       w2.GridReorderRequest{RecID: 2, MoveBefore: 9},
				ExpectedError: true,
			},
			{
				Input:         []int{},
				Request:       w2.GridReorderRequest{RecID: 1, MoveBefore: 2},
				ExpectedError: true,
			},
		}

		for _, test := range tests {
			sortedArray := make([]int, len(test.Input))
			copy(sortedArray, test.Input)

			err := w2sort.ReorderArray(sortedArray, test.Request)
			if err != nil {
				if !test.ExpectedError {
					t.Errorf("❌ Unexpected error for:\n  input: %v, %+v\n  error: %v",
						test.Input, test.Request, err)
				}
				continue
			}

			if test.ExpectedError {
				t.Errorf("❌ Expected error, but got none for:\n  input: %v, %+v",
					test.Input, test.Request)
				continue
			}

			if !slices.Equal(sortedArray, test.Expected) {
				t.Errorf("❌ Unexpected result for:\n  input: %v, %+v\n  got:   %v\n  want:  %v",
					test.Input, test.Request, sortedArray, test.Expected)
			}
		}
	})
}
