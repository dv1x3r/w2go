package w2sort

import (
	"fmt"

	"github.com/dv1x3r/w2go/w2"
)

func ReorderArray(a []int, r w2.GridReorderRequest) error {
	n := len(a)

	if n == 0 {
		return fmt.Errorf("slice is empty")
	}

	// find indexes
	iValue, iBefore := -1, -1
	for i, v := range a {
		if v == r.RecID {
			iValue = i
		}
		if !r.Bottom && v == r.MoveBefore {
			iBefore = i
		}
	}

	if iValue < 0 {
		return fmt.Errorf("id %d not found in slice", r.RecID)
	}

	if r.Bottom {
		iBefore = n
	} else if iBefore < 0 {
		return fmt.Errorf("moveBefore %d not found in slice", r.MoveBefore)
	}

	// already in right spot
	if r.Bottom {
		if iValue == n-1 {
			return nil
		}
	} else {
		if iValue+1 == iBefore {
			return nil
		}
	}

	// store the value to be moved
	tmp := a[iValue]

	if iValue < iBefore {
		// moving forward (value is before moveBefore)
		for i := iValue; i < iBefore-1; i++ {
			a[i] = a[i+1]
		}
		a[iBefore-1] = tmp
	} else if iValue > iBefore {
		// moving backward (value is after moveBefore)
		for i := iValue; i > iBefore; i-- {
			a[i] = a[i-1]
		}
		a[iBefore] = tmp
	}

	return nil
}
