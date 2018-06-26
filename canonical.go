package rush

type Canonicalizer struct {
	board     *Board
	memo      *Memo
	bestKey   MemoKey
	bestBoard *Board
}

func NewCanonicalizer(board *Board) *Canonicalizer {
	return &Canonicalizer{board, NewMemo(), board.memoKey, board.Copy()}
}

func (c *Canonicalizer) search(previousPiece int) {
	board := c.board

	if !c.memo.Add(board.MemoKey(), 0) {
		return
	}

	if board.memoKey.Less(&c.bestKey) {
		c.bestKey = board.memoKey
		c.bestBoard = board.Copy()
	}

	for _, move := range board.Moves(nil) {
		if move.Piece == previousPiece {
			continue
		}
		board.DoMove(move)
		c.search(move.Piece)
		board.UndoMove(move)
	}
}

func (c *Canonicalizer) Canonicalize() *Board {
	c.search(-1)
	c.bestBoard.SortPieces()
	return c.bestBoard
}
