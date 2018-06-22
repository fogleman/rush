package rush

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

*/

func blockedSquares(n int, positions, sizes []int, blocked []bool) []bool {
	result := make([]bool, n)
	return result
}
