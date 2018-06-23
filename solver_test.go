package rush

import "testing"

func TestStaticAnalysis(t *testing.T) {
	// ..xAA. => ....x.
	blockedSquaresForRow(6, []int{3}, []int{2}, []int{2})

	// AAA.BB => .xx.x.
	blockedSquaresForRow(6, []int{0, 4}, []int{3, 2}, []int{})

	// AA..BB => ......
	blockedSquaresForRow(6, []int{0, 4}, []int{2, 2}, []int{})

	// .x.AA. => ......
	blockedSquaresForRow(6, []int{3}, []int{2}, []int{1})

	// .xAA..BBx.. => ...........
	blockedSquaresForRow(11, []int{2, 6}, []int{2, 2}, []int{1, 8})

	// .xAAA.BBx.. => ...xx.x....
	blockedSquaresForRow(11, []int{2, 6}, []int{3, 2}, []int{1, 8})
}
