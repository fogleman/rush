package rush

type Unsolver struct {
	board     *Board
	solver    *Solver
	memo      *Memo
	bestDepth int
	bestBoard *Board
}

func NewUnsolver(board *Board) *Unsolver {
	solver := NewSolver(board)
	return &Unsolver{board, solver, NewMemo(), 0, board}
}

func (unsolver *Unsolver) search(depth, numMoves int) {
	board := unsolver.board

	if !unsolver.memo.Add(board.MemoKey(), 0) {
		return
	}

	if depth > unsolver.bestDepth {
		unsolver.bestDepth = depth
		unsolver.bestBoard = board.Copy()
		// fmt.Println(depth, numMoves)
		// fmt.Println(board)
		// fmt.Println(unsolver.memo.Size(), unsolver.memo.Hits())
		// fmt.Println()
	}

	for _, move := range board.Moves(nil) {
		board.DoMove(move)
		newNumMoves := unsolver.solver.solve(true).NumMoves
		delta := newNumMoves - numMoves
		if delta >= 0 {
			unsolver.search(depth+delta, newNumMoves)
		}
		board.UndoMove(move)
	}
}

func (unsolver *Unsolver) Unsolve() *Board {
	solution := unsolver.solver.Solve()
	if !solution.Solvable {
		return unsolver.board
	}
	unsolver.search(0, solution.NumMoves)
	return unsolver.bestBoard
}
