package rush

import "sort"

type Solver struct {
	board  *Board
	target int
	memo   *Memo
	path   []Move
	moves  [][]Move
}

type Solution struct {
	Solvable bool
	Moves    []Move
	NumMoves int
	NumSteps int
	Depth    int
	MemoSize int
	MemoHits uint64
}

func NewSolver(board *Board) *Solver {
	solver := Solver{}
	solver.board = board
	solver.target = board.Target()
	solver.memo = NewMemo()
	return &solver
}

func (solver *Solver) isSolved() bool {
	return solver.board.Pieces[0].Position == solver.target
}

func (solver *Solver) search(depth, maxDepth int) bool {
	height := maxDepth - depth
	if height == 0 {
		return solver.isSolved()
	}

	board := solver.board
	if !solver.memo.Add(board.MemoKey(), height) {
		return false
	}

	// count occupied squares between primary piece and target
	primary := board.Pieces[0]
	i0 := primary.Position + primary.Size
	i1 := solver.target + primary.Size - 1
	minMoves := 0
	for i := i0; i <= i1; i++ {
		if board.occupied[i] {
			minMoves++
		}
	}
	if minMoves >= height {
		return false
	}

	buf := &solver.moves[depth]
	*buf = board.Moves(*buf)
	for _, move := range *buf {
		board.DoMove(move)
		solved := solver.search(depth+1, maxDepth)
		board.UndoMove(move)
		if solved {
			solver.path[depth] = move
			return true
		}
	}
	return false
}

func (solver *Solver) Solve() Solution {
	board := solver.board
	memo := solver.memo

	if err := board.Validate(); err != nil {
		return Solution{}
	}

	if solver.isSolved() {
		return Solution{Solvable: true}
	}

	if board.Blocked() {
		return Solution{}
	}

	previousMemoSize := 0
	cutoff := board.Width - board.Pieces[0].Size
	for i := 1; ; i++ {
		solver.path = make([]Move, i)
		solver.moves = make([][]Move, i)
		if solver.search(0, i) {
			moves := solver.path
			steps := 0
			for _, move := range moves {
				steps += move.AbsSteps()
			}
			return Solution{
				Solvable: true,
				Moves:    moves,
				NumMoves: len(moves),
				NumSteps: steps,
				Depth:    i,
				MemoSize: memo.Size(),
				MemoHits: memo.Hits(),
			}
		}
		memoSize := memo.Size()
		if i > cutoff && memoSize == previousMemoSize {
			return Solution{
				Depth:    i,
				MemoSize: memo.Size(),
				MemoHits: memo.Hits(),
			}
		}
		previousMemoSize = memoSize
	}
}

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

012345	  012345
..xAA. => ....x.

n = 6 			# number of squares on the row (or column)
blocked = [2]   # blocked squares from the perpendicular orientation
positions = [3] # positions of pieces
sizes = [2]	 	# sizes of pieces
result = [4]	# blocked squares found by the algorithm

*/

func blockedSquaresForRow(w int, positions, sizes, blocked []int) []int {
	// for each piece, determine its range based on w and blocked
	n := len(positions)
	rs := make([][]int, n)
	lens := make([]int, n)
	for i := 0; i < n; i++ {
		p := positions[i]
		s := sizes[i]
		x0 := 0
		x1 := w - s
		for _, b := range blocked {
			if b < p {
				x0 = maxInt(x0, b+1)
			}
			if b > p {
				x1 = minInt(x1, b-s)
			}
		}
		d := x1 - x0 + 1
		r := make([]int, d)
		for j := 0; j < d; j++ {
			r[j] = x0 + j
		}
		rs[i] = r
		lens[i] = len(r)
	}
	// do something like itertools.product in python
	count := 0
	counts := make([]int, w)
	idx := make([]int, n)
	for {
		// make sure pieces aren't overlapping
		ok := true
		for i := 1; i < n; i++ {
			j := i - 1
			if rs[i][idx[i]]-rs[j][idx[j]] < sizes[j] {
				ok = false
				break
			}
		}
		if ok {
			// increment count
			count++
			// increment counts for occupied squares
			for i := 0; i < n; i++ {
				p := rs[i][idx[i]]
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
	var result []int
	for i, n := range counts {
		if n == count {
			result = append(result, i)
		}
	}
	// fmt.Println(count, counts, result)
	return result
}

func updateBlocked(board *Board, horz, vert []bool) bool {
	changed := false
	w := board.Width
	h := board.Height
	// copy and sort pieces by position
	pieces := make([]Piece, len(board.Pieces))
	copy(pieces, board.Pieces)
	sort.Slice(pieces, func(i, j int) bool {
		return pieces[i].Position < pieces[j].Position
	})
	// iterate over rows
	for y := 0; y < h; y++ {
		var positions, sizes, blocked []int
		for _, piece := range pieces {
			if piece.Orientation != Horizontal {
				continue
			}
			if piece.Row(w) != y {
				continue
			}
			positions = append(positions, piece.Col(w))
			sizes = append(sizes, piece.Size)
		}
		if len(positions) == 0 {
			continue
		}
		i0 := y * w
		for i := 0; i < w; i++ {
			if horz[i0+i] {
				blocked = append(blocked, i)
			}
		}
		result := blockedSquaresForRow(w, positions, sizes, blocked)
		for _, i := range result {
			i += i0
			if !vert[i] {
				vert[i] = true
				changed = true
			}
		}
	}
	// iterate over cols
	for x := 0; x < w; x++ {
		var positions, sizes, blocked []int
		for _, piece := range pieces {
			if piece.Orientation != Vertical {
				continue
			}
			if piece.Col(w) != x {
				continue
			}
			positions = append(positions, piece.Row(w))
			sizes = append(sizes, piece.Size)
		}
		if len(positions) == 0 {
			continue
		}
		i0 := x
		for i := 0; i < h; i++ {
			if vert[i0+i*w] {
				blocked = append(blocked, i)
			}
		}
		result := blockedSquaresForRow(h, positions, sizes, blocked)
		for _, i := range result {
			i = i*w + x
			if !horz[i] {
				horz[i] = true
				changed = true
			}
		}
	}
	return changed
}

func targetIsBlocked(board *Board) bool {
	w := board.Width
	h := board.Height
	horz := make([]bool, w*h)
	vert := make([]bool, w*h)
	for updateBlocked(board, horz, vert) {
	}
	piece := board.Pieces[0]
	i0 := piece.Position + piece.Size
	i1 := (piece.Row(w) + 1) * w
	for i := i0; i < i1; i++ {
		if horz[i] || vert[i] {
			return true
		}
	}
	return false
}
