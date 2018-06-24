package rush

type Unsolver struct {
	board        *Board
	solver       *Solver
	memo         *Memo
	bestNumMoves int
	bestBoard    *Board
}

func NewUnsolver(board *Board) *Unsolver {
	solver := NewSolver(board)
	memo := NewMemo()
	return &Unsolver{board, solver, memo, 0, board}
}

func (unsolver *Unsolver) search(numMoves int) {
	board := unsolver.board

	if !unsolver.memo.Add(board.MemoKey(), 0) {
		return
	}

	if numMoves > unsolver.bestNumMoves {
		unsolver.bestNumMoves = numMoves
		unsolver.bestBoard = board.Copy()
	}

	for _, move := range board.Moves(nil) {
		board.DoMove(move)
		newNumMoves := unsolver.solver.solve(true).NumMoves
		if newNumMoves-numMoves >= 0 {
			unsolver.search(newNumMoves)
		}
		board.UndoMove(move)
	}
}

func (unsolver *Unsolver) Unsolve() *Board {
	solution := unsolver.solver.Solve()
	if !solution.Solvable {
		return unsolver.bestBoard
	}
	unsolver.search(solution.NumMoves)
	return unsolver.bestBoard
}
