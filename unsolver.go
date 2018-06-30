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
	delta := solution.NumMoves - u.bestSolution.NumMoves
	if delta > 0 || (delta == 0 && board.MemoKey().Less(u.bestBoard.MemoKey(), true)) {
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

func (u *Unsolver) Unsolve() (*Board, Solution) {
	u.bestBoard = u.board
	u.bestSolution = u.solver.Solve()
	if u.bestSolution.Solvable {
		u.search(-1)
	}
	return u.bestBoard, u.bestSolution
}
