package rush

import (
	"reflect"
	"testing"
)

func TestBlockedSquares(t *testing.T) {
	test := func(w int, positions, sizes, blocked, expected []int) {
		result := theStaticAnalyzer.blockedSquares(w, positions, sizes, blocked)
		if !reflect.DeepEqual(result, expected) {
			t.Fail()
		}
	}

	// ..xAA. => ....x.
	test(6, []int{3}, []int{2}, []int{2}, []int{4})

	// AAA.BB => .xx.x.
	test(6, []int{0, 4}, []int{3, 2}, []int{}, []int{1, 2, 4})

	// AA..BB => ......
	test(6, []int{0, 4}, []int{2, 2}, []int{}, []int{})

	// .x.AA. => ......
	test(6, []int{3}, []int{2}, []int{1}, []int{})

	// .xAA..BBx.. => ...........
	test(11, []int{2, 6}, []int{2, 2}, []int{1, 8}, []int{})

	// .xAAA.BBx.. => ...xx.x....
	test(11, []int{2, 6}, []int{3, 2}, []int{1, 8}, []int{3, 4, 6})
}
