package rush

import "testing"

func TestStaticAnalysis(t *testing.T) {
	f := theStaticAnalyzer.blockedSquares

	// ..xAA. => ....x.
	f(6, []int{3}, []int{2}, []int{2})

	// AAA.BB => .xx.x.
	f(6, []int{0, 4}, []int{3, 2}, []int{})

	// AA..BB => ......
	f(6, []int{0, 4}, []int{2, 2}, []int{})

	// .x.AA. => ......
	f(6, []int{3}, []int{2}, []int{1})

	// .xAA..BBx.. => ...........
	f(11, []int{2, 6}, []int{2, 2}, []int{1, 8})

	// .xAAA.BBx.. => ...xx.x....
	f(11, []int{2, 6}, []int{3, 2}, []int{1, 8})
}
