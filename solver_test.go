package rush

import "testing"

func TestStaticAnalysis(t *testing.T) {
	// ..xAA. => ....x.
	blockedSquares(6, []int{3}, []int{2}, []int{2})

	// AAA.BB => .xx.x.
	blockedSquares(6, []int{0, 4}, []int{3, 2}, []int{})

	// AA..BB => ......
	blockedSquares(6, []int{0, 4}, []int{2, 2}, []int{})

	// .x.AA. => ......
	blockedSquares(6, []int{3}, []int{2}, []int{1})

	// .xAA..BBx.. => ...........
	blockedSquares(11, []int{2, 6}, []int{2, 2}, []int{1, 8})

	// .xAAA.BBx.. => ...xx.x....
	blockedSquares(11, []int{2, 6}, []int{3, 2}, []int{1, 8})
}
