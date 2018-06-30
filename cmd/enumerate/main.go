package main

import (
	"fmt"

	. "github.com/fogleman/rush"
)

const (
	// W  = 4
	// H  = 4
	// Py = 1
	// Px = 2

	// W  = 5
	// H  = 5
	// Py = 2
	// Px = 3

	W  = 6
	H  = 6
	Py = 2
	Px = 4

	P = Py*W + Px
)

type Enumerator struct {
	Board           *Board
	Seen            map[string]bool
	Memo            *Memo
	Solver          *Solver
	HardestSolution Solution
	HardestBoard    *Board
	Canonical       bool
	CanonicalKey    MemoKey
	Count           int
}

func NewEnumerator(board *Board) *Enumerator {
	e := &Enumerator{}
	e.Board = board
	e.Seen = make(map[string]bool)
	return e
}

func (e *Enumerator) hardestSearch(previousPiece int) {
	board := e.Board

	if !e.Memo.Add(board.MemoKey(), 0) {
		return
	}

	solution := e.Solver.UnsafeSolve()
	delta := solution.NumMoves - e.HardestSolution.NumMoves
	if delta > 0 || (delta == 0 && board.MemoKey().Less(e.HardestBoard.MemoKey(), true)) {
		e.HardestSolution = solution
		e.HardestBoard = board.Copy()
	}

	for _, move := range board.Moves(nil) {
		if move.Piece == previousPiece {
			continue
		}
		board.DoMove(move)
		e.hardestSearch(move.Piece)
		board.UndoMove(move)
	}
}

func (e *Enumerator) HardestSearch() {
	e.Memo = NewMemo()
	e.Solver = NewSolver(e.Board)
	e.HardestBoard = e.Board.Copy()
	e.HardestSolution = e.Solver.Solve()
	e.hardestSearch(-1)
	e.HardestBoard.SortPieces()
}

func (e *Enumerator) canonicalSearch(previousPiece int) {
	if !e.Canonical {
		return
	}

	board := e.Board

	if !e.Memo.Add(board.MemoKey(), 0) {
		return
	}

	if board.MemoKey().Less(&e.CanonicalKey, false) {
		e.Canonical = false
		return
	}

	for _, move := range board.Moves(nil) {
		if move.Piece == 0 {
			continue
		}
		if move.Piece == previousPiece {
			continue
		}
		board.DoMove(move)
		e.canonicalSearch(move.Piece)
		board.UndoMove(move)
	}
}

func (e *Enumerator) CanonicalSearch() {
	e.Memo = NewMemo()
	e.Canonical = true
	e.CanonicalKey = *e.Board.MemoKey()
	e.canonicalSearch(-1)
}

func (e *Enumerator) place(after int) {
	board := e.Board

	if board.HasFullRowOrCol() {
		return
	}

	e.CanonicalSearch()
	if !e.Canonical {
		return
	}

	e.HardestSearch()
	hardest := e.HardestBoard
	solution := e.HardestSolution

	if solution.NumMoves == 0 {
		return
	}

	if solution.NumMoves >= 1 {
		key := hardest.Hash()
		_, seen := e.Seen[key]
		if !seen {
			e.Seen[key] = true
			fmt.Printf("%02d %02d %s %d\n", solution.NumMoves, solution.NumSteps, key, solution.MemoSize)
		} else {
			e.Count++
		}
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
				if o == Horizontal && y == Py {
					continue
				}
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
	board.AddPiece(Piece{P, 2, Horizontal})
	e := NewEnumerator(board)
	e.Enumerate()
	fmt.Println(len(e.Seen), e.Count)
}
