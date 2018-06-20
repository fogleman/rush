package rush

type Solver struct {
	Board  *Board
	Target int
	memo   *Memo
	path   []Move
	moves  [][]Move
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

func (solver *Solver) Solve() ([]Move, bool) {
	if solver.isSolved() {
		return nil, true
	}
	previousMemoSize := 0
	for i := 1; ; i++ {
		solver.path = make([]Move, i)
		solver.moves = make([][]Move, i)
		if solver.search(0, i) {
			return solver.path, true
		}
		memoSize := solver.memo.Size()
		if memoSize == previousMemoSize {
			return nil, false
		}
		previousMemoSize = memoSize
	}
}
