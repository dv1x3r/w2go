package w2sort_test

import (
	"slices"
	"testing"

	"github.com/dv1x3r/w2go/w2sort"
)

func TestReorderArray(t *testing.T) {
	t.Run("ReorderArray", func(t *testing.T) {
		tests := []struct {
			Input         []int
			ID            int
			MoveBefore    int
			Last          bool
			Expected      []int
			ExpectedError bool
		}{
			{
				Input:      []int{1, 2, 3, 4, 5},
				ID:         2,
				MoveBefore: 4,
				Expected:   []int{1, 3, 2, 4, 5},
			},
			{
				Input:      []int{1, 2, 3, 4, 5},
				ID:         4,
				MoveBefore: 1,
				Expected:   []int{4, 1, 2, 3, 5},
			},
			{
				Input:    []int{1, 2, 3, 4, 5},
				ID:       3,
				Last:     true,
				Expected: []int{1, 2, 4, 5, 3},
			},
			{
				Input:         []int{1, 2, 3},
				ID:            9,
				MoveBefore:    1,
				ExpectedError: true,
			},
			{
				Input:         []int{1, 2, 3},
				ID:            2,
				MoveBefore:    9,
				ExpectedError: true,
			},
			{
				Input:         []int{},
				ID:            1,
				MoveBefore:    2,
				ExpectedError: true,
			},
		}

		for _, test := range tests {
			sortedArray := make([]int, len(test.Input))
			copy(sortedArray, test.Input)

			err := w2sort.ReorderArray(sortedArray, test.ID, test.MoveBefore, test.Last)
			if err != nil {
				if !test.ExpectedError {
					t.Errorf("❌ Unexpected error for:\n  input: %v, id: %v, moveBefore: %v, last: %v\n  error: %v",
						test.Input, test.ID, test.MoveBefore, test.Last, err)
				}
				continue
			}

			if test.ExpectedError {
				t.Errorf("❌ Expected error, but got none for:\n  input: %v, id: %v, moveBefore: %v, last: %v",
					test.Input, test.ID, test.MoveBefore, test.Last)
				continue
			}

			if !slices.Equal(sortedArray, test.Expected) {
				t.Errorf("❌ Unexpected result for:\n  input: %v, id: %v, moveBefore: %v, last: %v\n  got:   %v\n  want:  %v",
					test.Input, test.ID, test.MoveBefore, test.Last, sortedArray, test.Expected)
			}
		}
	})
}
