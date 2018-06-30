package main

import (
	"fmt"
	"runtime"

	. "github.com/fogleman/rush"
)

const (
	W = 6
	H = 6

	PrimaryRow  = 2
	PrimarySize = 2

	MinSize = 2
	MaxSize = 3
)

// Enumerator generates all possible Boards
type Enumerator struct {
	board      *Board
	primaryRow int
	minSize    int
	maxSize    int
	ch         chan *Board
}

func NewEnumerator(w, h, pr, ps, mins, maxs int) *Enumerator {
	target := (PrimaryRow+1)*w - PrimarySize
	board := NewEmptyBoard(w, h)
	board.AddPiece(Piece{target, PrimarySize, Horizontal})
	ch := make(chan *Board, 1024)
	return &Enumerator{board, pr, mins, maxs, ch}
}

func (e *Enumerator) Enumerate() <-chan *Board {
	go func() {
		e.place(-1)
		close(e.ch)
	}()
	return e.ch
}

func (e *Enumerator) place(after int) {
	board := e.board
	if board.HasFullRowOrCol() {
		return
	}
	e.ch <- board.Copy()
	for o := Horizontal; o <= Vertical; o++ {
		for s := e.minSize; s <= e.maxSize; s++ {
			xx := board.Width
			yy := board.Height
			if o == Horizontal {
				xx = W - s + 1
			} else {
				yy = H - s + 1
			}
			for y := 0; y < yy; y++ {
				if o == Horizontal && y == e.primaryRow {
					continue
				}
				for x := 0; x < xx; x++ {
					p := y*W + x
					if p <= after {
						continue
					}
					if board.AddPiece(Piece{p, s, o}) {
						e.place(p)
						board.RemoveLastPiece()
					}
				}
			}
		}
	}
}

func isCanonical(board *Board, memo *Memo, key *MemoKey, previousPiece int) bool {
	if board.MemoKey().Less(key, false) {
		return false
	}
	if !memo.Add(board.MemoKey(), 0) {
		return true
	}
	for _, move := range board.Moves(nil) {
		if move.Piece == 0 {
			continue
		}
		if move.Piece == previousPiece {
			continue
		}
		board.DoMove(move)
		ok := isCanonical(board, memo, key, move.Piece)
		board.UndoMove(move)
		if !ok {
			return false
		}
	}
	return true
}

func IsCanonical(board *Board) bool {
	memo := NewMemo()
	key := *board.MemoKey()
	return isCanonical(board, memo, &key, -1)
}

type Result struct {
	Board    *Board
	Solution Solution
	Done     bool
}

func worker(boards <-chan *Board, results chan<- Result) {
	sa := NewStaticAnalyzer()
	for board := range boards {
		if !IsCanonical(board) {
			continue
		}
		unsolver := NewUnsolverWithStaticAnalyzer(board, sa)
		unsolved, solution := unsolver.Unsolve()
		unsolved.SortPieces()
		if solution.NumMoves >= 1 {
			results <- Result{unsolved, solution, false}
		}
	}
	results <- Result{Done: true}
}

func main() {
	e := NewEnumerator(W, H, PrimaryRow, PrimarySize, MinSize, MaxSize)
	boards := e.Enumerate()
	results := make(chan Result, 1024)

	wn := runtime.NumCPU()
	for i := 0; i < wn; i++ {
		go worker(boards, results)
	}

	seen := make(map[string]bool)
	for result := range results {
		if result.Done {
			wn--
			if wn == 0 {
				break
			}
			continue
		}
		unsolved := result.Board
		solution := result.Solution
		key := unsolved.Hash()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = true
		fmt.Printf(
			"%02d %02d %s %d\n",
			solution.NumMoves, solution.NumSteps, key, solution.MemoSize)
	}
}
