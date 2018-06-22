package rush

type Solver struct {
	Board  *Board
	Target int
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

func NewSolver(board *Board, target int) *Solver {
	memo := NewMemo()
	return &Solver{board, target, memo, nil, nil}
}

func (solver *Solver) isSolved() bool {
	return solver.Board.Pieces[0].Position == solver.Target
}

func (solver *Solver) search(depth, maxDepth int) bool {
	if solver.isSolved() {
		return true
	}
	if depth == maxDepth {
		return false
	}
	board := solver.Board
	height := maxDepth - depth
	if !solver.memo.Add(board.MemoKey(), height) {
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
	if solver.isSolved() {
		return Solution{Solvable: true}
	}
	if !solver.sanityCheck() {
		return Solution{}
	}
	previousMemoSize := 0
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
				MemoSize: solver.memo.Size(),
				MemoHits: solver.memo.Hits(),
			}
		}
		memoSize := solver.memo.Size()
		if memoSize == previousMemoSize {
			return Solution{
				Depth:    i,
				MemoSize: solver.memo.Size(),
				MemoHits: solver.memo.Hits(),
			}
		}
		previousMemoSize = memoSize
	}
}

// sanityCheck performs basic tests to see if the puzzle is obviously not
// solvable, returning false if this is the case
func (solver *Solver) sanityCheck() bool {
	target := solver.Target
	board := solver.Board
	w := board.Width
	h := board.Height
	pieces := board.Pieces
	primary := pieces[0]
	// check aligned pieces on same row or column as the primary piece
	var before, after int
	for _, piece := range pieces[1:] {
		if piece.Orientation != primary.Orientation {
			continue
		}
		if piece.Orientation == Horizontal {
			if piece.Row(w) != primary.Row(w) {
				continue
			}
		} else {
			if piece.Col(w) != primary.Col(w) {
				continue
			}
		}
		if piece.Position < primary.Position {
			before += piece.Size
		} else {
			after += piece.Size
		}
	}
	var i0, i1 int
	if primary.Orientation == Horizontal {
		i0 = primary.Row(w) * w
		i1 = i0 + w - primary.Size
		i0 += before
		i1 -= after
	} else {
		i0 = primary.Col(w)
		i1 = i0 + (h-primary.Size)*w
		i0 += before * w
		i1 -= after * w
	}
	if target < i0 || target > i1 {
		return false
	}
	return true
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

*/

func blockedSquares(n int, positions, sizes []int, blocked []bool) []bool {
	result := make([]bool, n)
	return result
}
