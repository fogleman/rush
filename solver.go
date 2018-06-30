package rush

type Solution struct {
	Solvable bool
	Moves    []Move
	NumMoves int
	NumSteps int
	Depth    int
	MemoSize int
	MemoHits uint64
}

type Solver struct {
	board  *Board
	target int
	memo   *Memo
	sa     *StaticAnalyzer
	path   []Move
	moves  [][]Move
}

func NewSolverWithStaticAnalyzer(board *Board, sa *StaticAnalyzer) *Solver {
	solver := Solver{}
	solver.board = board
	solver.target = board.Target()
	solver.memo = NewMemo()
	solver.sa = sa
	return &solver
}

func NewSolver(board *Board) *Solver {
	return NewSolverWithStaticAnalyzer(board, theStaticAnalyzer)
}

func (solver *Solver) isSolved() bool {
	return solver.board.Pieces[0].Position == solver.target
}

func (solver *Solver) search(depth, maxDepth, previousPiece int) bool {
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
		if move.Piece == previousPiece {
			continue
		}
		board.DoMove(move)
		solved := solver.search(depth+1, maxDepth, move.Piece)
		board.UndoMove(move)
		if solved {
			solver.memo.Set(board.MemoKey(), height-1)
			solver.path[depth] = move
			return true
		}
	}
	return false
}

func (solver *Solver) solve(skipChecks bool) Solution {
	board := solver.board
	memo := solver.memo

	if !skipChecks {
		if err := board.Validate(); err != nil {
			return Solution{}
		}
		if solver.sa.Impossible(board) {
			return Solution{}
		}
	}

	if solver.isSolved() {
		return Solution{Solvable: true}
	}

	previousMemoSize := 0
	noChange := 0
	cutoff := board.Width - board.Pieces[0].Size
	for i := 1; ; i++ {
		solver.path = make([]Move, i)
		solver.moves = make([][]Move, i)
		if solver.search(0, i, -1) {
			moves := solver.path
			steps := 0
			for _, move := range moves {
				steps += move.AbsSteps()
			}
			result := Solution{
				Solvable: true,
				Moves:    moves,
				NumMoves: len(moves),
				NumSteps: steps,
				Depth:    i,
				MemoSize: memo.Size(),
				MemoHits: memo.Hits(),
			}
			return result
		}
		memoSize := memo.Size()
		if memoSize == previousMemoSize {
			noChange++
		} else {
			noChange = 0
		}
		if !skipChecks && noChange > cutoff {
			return Solution{
				Depth:    i,
				MemoSize: memo.Size(),
				MemoHits: memo.Hits(),
			}
		}
		previousMemoSize = memoSize
	}
}

func (solver *Solver) Solve() Solution {
	return solver.solve(false)
}

func (solver *Solver) UnsafeSolve() Solution {
	return solver.solve(true)
}
