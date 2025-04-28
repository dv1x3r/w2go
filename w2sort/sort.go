package w2sort

import "fmt"

func ReorderArray(a []int, id int, moveBefore int, last bool) error {
	n := len(a)

	if n == 0 {
		return fmt.Errorf("slice is empty")
	}

	// find indexes
	iValue, iBefore := -1, -1
	for i, v := range a {
		if v == id {
			iValue = i
		}
		if !last && v == moveBefore {
			iBefore = i
		}
	}

	if iValue < 0 {
		return fmt.Errorf("id %d not found in slice", id)
	}

	if last {
		iBefore = n
	} else if iBefore < 0 {
		return fmt.Errorf("moveBefore %d not found in slice", moveBefore)
	}

	// already in right spot
	if last {
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
