package rush

/*

Static analysis code is below. Its purpose is to detect if a Board will be
impossible to solve without actually doing an expensive recursive search.
Certain patterns, frequent among randomly generated boards, can be relatively
easily detected and weeded out as impossible to solve.

Consider the following row. We will analyze it in isolation.

AAA.BB

There are only three possible layouts for these two pieces:

AAABB.
AAA.BB
.AAABB

Of the six squares on this row, three of them are always occupied no matter
the configuration of the pieces:

.xx.x.

We will call these squares "blocked."

We can examine all rows and columns on the board for such "blocked" squares.

If any of the squares between the primary piece (the "red car") and its exit
are blocked, then we know that the puzzle cannot be solved.

But that's not all! Blocked squares on a row affect the possibilities on the
intersecting columns. Let's take the blocked squares from above and consider
an example column:

  .
  .
.xx.x.
  C
  C
  .

Without considering blocked squares, it seems that the C piece could
potentially traverse the entire column. Actually, the C piece will be
constrained to the bottom two squares in the column, making the second from
the bottom square also blocked:

  .
  .
.xx.x.
  .
  x
  .

We can repeat this process of identifying blocked squares based on each row
and column's configuration and existing blocked squares until no new squares
are identified.

So we need an algorithm that can take the pieces present on a row or column,
along with already-identified blocked squares from the perpendicular direction,
and return a set of blocked squares. Returning to the column above:

..xCC. => ....x.

Note that we need to distinguish between horizontally blocked squares and
vertically blocked squares. So the example above does not retain blocked
squares from its input. Here are some example inputs and outputs using our
ASCII based representation:

..xAA. => ....x.
AAA.BB => .xx.x.
AA..BB => ......
.x.AA. => ......

Let's figure out the appropriate data structures. We'll use this example:

012345    012345
..xAA. => ....x.

n = 6           # number of squares on the row (or column)
blocked = [2]   # blocked squares from the perpendicular orientation
positions = [3] # positions of pieces
sizes = [2]     # sizes of pieces
result = [4]    # blocked squares found by the algorithm

*/

var theStaticAnalyzer = NewStaticAnalyzer()

type StaticAnalyzer struct {
	// these buffers are allocated once so multiple static analyses can be
	// performed faster (less GC)
	horz       []bool
	vert       []bool
	positions  []int
	sizes      []int
	blocked    []int
	lens       []int
	idx        []int
	counts     []int
	result     []int
	placements [][]int
}

func NewStaticAnalyzer() *StaticAnalyzer {
	maxPiecesPerRow := MaxBoardSize / MinPieceSize
	maxPlacementsPerRow := MaxBoardSize - MinPieceSize + 1
	sa := &StaticAnalyzer{}
	sa.horz = make([]bool, MaxBoardSize*MaxBoardSize)
	sa.vert = make([]bool, MaxBoardSize*MaxBoardSize)
	sa.positions = make([]int, maxPiecesPerRow)
	sa.sizes = make([]int, maxPiecesPerRow)
	sa.blocked = make([]int, MaxBoardSize)
	sa.lens = make([]int, maxPiecesPerRow)
	sa.idx = make([]int, maxPiecesPerRow)
	sa.counts = make([]int, MaxBoardSize)
	sa.result = make([]int, MaxBoardSize)
	sa.placements = make([][]int, maxPiecesPerRow)
	for i := range sa.placements {
		sa.placements[i] = make([]int, maxPlacementsPerRow)
	}
	return sa
}

func (sa *StaticAnalyzer) Impossible(board *Board) bool {
	// run analysis
	sa.analyze(board)
	// see if any squares between the primary piece and its exit are blocked
	w := board.Width
	piece := board.Pieces[0]
	i0 := piece.Position + piece.Size
	i1 := (piece.Row(w) + 1) * w
	for i := i0; i < i1; i++ {
		if sa.horz[i] || sa.vert[i] {
			return true
		}
	}
	return false
}

func (sa *StaticAnalyzer) BlockedSquares(board *Board) []int {
	// run analysis
	sa.analyze(board)
	// compile a list of all blocked squares
	n := board.Width * board.Height
	var result []int
	for i := 0; i < n; i++ {
		if sa.horz[i] || sa.vert[i] {
			result = append(result, i)
		}
	}
	return result
}

func (sa *StaticAnalyzer) analyze(board *Board) {
	// zero out buffers
	for i := range sa.horz {
		sa.horz[i] = false
		sa.vert[i] = false
	}
	// walls are always blocked for both directions
	for _, i := range board.Walls {
		sa.horz[i] = true
		sa.vert[i] = true
	}
	// run the step function until no more changes are made
	for sa.step(board) {
	}
}

func (sa *StaticAnalyzer) step(board *Board) bool {
	changed := false
	w := board.Width
	h := board.Height
	pieces := board.Pieces
	// iterate over rows
	for y := 0; y < h; y++ {
		// find all pieces in this row
		positions, sizes := sa.positions[:0], sa.sizes[:0]
		for _, piece := range pieces {
			if piece.Orientation == Horizontal && piece.Row(w) == y {
				positions = append(positions, piece.Col(w))
				sizes = append(sizes, piece.Size)
			}
		}
		// abort early if row is empty
		if len(positions) == 0 {
			continue
		}
		// figure out which squares are blocked from opposite direction
		blocked := sa.blocked[:0]
		i0 := y * w
		for i := 0; i < w; i++ {
			if sa.horz[i0+i] {
				blocked = append(blocked, i)
			}
		}
		// update blocked squares on this row
		result := sa.blockedSquares(w, positions, sizes, blocked)
		for _, i := range result {
			i = i + i0
			if !sa.vert[i] {
				sa.vert[i] = true
				changed = true
			}
		}
	}
	// iterate over cols
	for x := 0; x < w; x++ {
		// find all pieces in this col
		positions, sizes := sa.positions[:0], sa.sizes[:0]
		for _, piece := range pieces {
			if piece.Orientation == Vertical && piece.Col(w) == x {
				positions = append(positions, piece.Row(w))
				sizes = append(sizes, piece.Size)
			}
		}
		// abort early if col is empty
		if len(positions) == 0 {
			continue
		}
		// figure out which squares are blocked from opposite direction
		blocked := sa.blocked[:0]
		i0 := x
		for i := 0; i < h; i++ {
			if sa.vert[i0+i*w] {
				blocked = append(blocked, i)
			}
		}
		// update blocked squares on this col
		result := sa.blockedSquares(h, positions, sizes, blocked)
		for _, i := range result {
			i = i*w + x
			if !sa.horz[i] {
				sa.horz[i] = true
				changed = true
			}
		}
	}
	// return true if any changes were made
	return changed
}

func (sa *StaticAnalyzer) blockedSquares(w int, positions, sizes, blocked []int) []int {
	n := len(positions)
	// insertion sort the positions & sizes together
	for i := 1; i < n; i++ {
		for j := i; j > 0 && positions[j] < positions[j-1]; j-- {
			positions[j], positions[j-1] = positions[j-1], positions[j]
			sizes[j], sizes[j-1] = sizes[j-1], sizes[j]
		}
	}
	// for each piece, determine its possible placements based on w and blocked
	placements := sa.placements
	lens := sa.lens[:n]
	for i := 0; i < n; i++ {
		p := positions[i]
		s := sizes[i]
		// init placement range to the full row
		x0 := 0
		x1 := w - s
		// reduce placement range based on surrounding blocked squares
		for _, b := range blocked {
			if b < p {
				x0 = maxInt(x0, b+1)
			}
			if b > p {
				x1 = minInt(x1, b-s)
			}
		}
		// make a list of all possible piece positions
		d := x1 - x0 + 1
		for j := 0; j < d; j++ {
			placements[i][j] = x0 + j
		}
		lens[i] = d
	}
	// do something like itertools.product in python to examine all possible
	// placements of all the pieces together
	count := 0
	// zero out reused buffers
	counts := sa.counts[:w]
	idx := sa.idx[:n]
	for i := range counts {
		counts[i] = 0
	}
	for i := range idx {
		idx[i] = 0
	}
	for {
		// make sure pieces aren't overlapping
		ok := true
		for i := 1; i < n; i++ {
			j := i - 1
			if placements[i][idx[i]]-placements[j][idx[j]] < sizes[j] {
				ok = false
				break
			}
		}
		if ok {
			// increment count
			count++
			// increment counts for occupied squares
			for i := 0; i < n; i++ {
				p := placements[i][idx[i]]
				s := sizes[i]
				for j := 0; j < s; j++ {
					counts[p+j]++
				}
			}
		}
		// go to next lexicographic index
		i := n - 1
		for ; i >= 0 && idx[i] == lens[i]-1; i-- {
			idx[i] = 0
		}
		if i < 0 {
			break
		}
		idx[i]++
	}
	// see which squares were always occupied
	result := sa.result[:0]
	for i, n := range counts {
		if n == count {
			result = append(result, i)
		}
	}
	return result
}
