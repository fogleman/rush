package main

import (
	"fmt"

	. "github.com/fogleman/rush"
)

const (
	W  = 4
	H  = 4
	PP = 6

// W  = 5
// H  = 5
// PP = 13
)

type Enumerator struct {
	Board        *Board
	Seen         map[string]bool
	Memo         *Memo
	Solver       *Solver
	BestSolution Solution
	BestBoard    *Board
	Count        int
}

func NewEnumerator(board *Board) *Enumerator {
	e := &Enumerator{}
	e.Board = board
	e.Seen = make(map[string]bool)
	return e
}

func (e *Enumerator) search(previousPiece int) {
	board := e.Board

	if !e.Memo.Add(board.MemoKey(), 0) {
		return
	}

	solution := e.Solver.UnsafeSolve()
	if solution.NumMoves >= e.BestSolution.NumMoves {
		if board.MemoKey().Less(e.BestBoard.MemoKey()) {
			e.BestSolution = solution
			e.BestBoard = board.Copy()
		}
	}

	for _, move := range board.Moves(nil) {
		if move.Piece == previousPiece {
			continue
		}
		board.DoMove(move)
		e.search(move.Piece)
		board.UndoMove(move)
	}
}

func (e *Enumerator) Search() {
	e.Memo = NewMemo()
	e.Solver = NewSolver(e.Board)
	e.BestBoard = e.Board.Copy()
	e.BestSolution = e.Solver.Solve()
	e.search(-1)
	e.BestBoard.SortPieces()
}

func (e *Enumerator) place(after int) {
	board := e.Board

	e.Search()
	canonical := e.BestBoard
	solution := e.BestSolution

	key := canonical.Hash()
	if _, ok := e.Seen[key]; ok {
		return
		// 1393 431725
	}
	e.Seen[key] = true

	if solution.NumMoves >= 2 {
		e.Count++
		fmt.Println(len(e.Seen))
		fmt.Println(solution.NumMoves)
		fmt.Println(canonical)
		fmt.Println()
	}

	w := board.Width
	h := board.Height
	i := len(board.Pieces)

	for o := Horizontal; o <= Vertical; o++ {
		for s := 2; s <= 3; s++ {
			xx := w
			yy := h
			if o == Horizontal {
				xx = W - s + 1
			} else {
				yy = H - s + 1
			}
			for y := 0; y < yy; y++ {
				for x := 0; x < xx; x++ {
					p := y*W + x
					if p <= after {
						continue
					}
					if board.AddPiece(Piece{p, s, o}) {
						e.place(p)
						board.RemovePiece(i)
					}
				}
			}
		}
	}
}

func (e *Enumerator) Enumerate() {
	e.place(-1)
}

func main() {
	board := NewEmptyBoard(W, H)
	board.AddPiece(Piece{PP, 2, Horizontal})
	e := NewEnumerator(board)
	e.Enumerate()
	fmt.Println(len(e.Seen), e.Count)
}
