package rush

type Unsolver struct {
	board        *Board
	solver       *Solver
	memo         *Memo
	bestBoard    *Board
	bestSolution Solution
}

func NewUnsolverWithStaticAnalyzer(board *Board, sa *StaticAnalyzer) *Unsolver {
	board = board.Copy()
	u := Unsolver{}
	u.board = board
	u.solver = NewSolverWithStaticAnalyzer(board, sa)
	u.memo = NewMemo()
	return &u
}

func NewUnsolver(board *Board) *Unsolver {
	return NewUnsolverWithStaticAnalyzer(board, theStaticAnalyzer)
}

func (u *Unsolver) search(previousPiece int) {
	board := u.board

	if !u.memo.Add(board.MemoKey(), 0) {
		return
	}

	solution := u.solver.UnsafeSolve()

	better := false
	dNumMoves := solution.NumMoves - u.bestSolution.NumMoves
	dNumSteps := solution.NumSteps - u.bestSolution.NumSteps
	if dNumMoves >= 0 {
		if dNumMoves > 0 {
			better = true
		} else if dNumMoves == 0 && dNumSteps > 0 {
			better = true
		} else if dNumMoves == 0 && dNumSteps == 0 && board.MemoKey().Less(u.bestBoard.MemoKey(), true) {
			better = true
		}
	}

	if better {
		u.bestSolution = solution
		u.bestBoard = board.Copy()
	}

	for _, move := range board.Moves(nil) {
		if move.Piece == previousPiece {
			continue
		}
		board.DoMove(move)
		u.search(move.Piece)
		board.UndoMove(move)
	}
}

func (u *Unsolver) unsolve(skipChecks bool) (*Board, Solution) {
	u.bestBoard = u.board.Copy()
	u.bestSolution = u.solver.solve(skipChecks)
	if u.bestSolution.Solvable {
		u.search(-1)
	}
	return u.bestBoard, u.bestSolution
}

func (u *Unsolver) Unsolve() (*Board, Solution) {
	return u.unsolve(false)
}

func (u *Unsolver) UnsafeUnsolve() (*Board, Solution) {
	return u.unsolve(true)
}
