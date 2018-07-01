package main

import (
	"fmt"
	"os"
	"runtime"
	"time"

	. "github.com/fogleman/rush"
)

const (
	W = 6
	H = 6

	PrimaryRow  = 2
	PrimarySize = 2

	MinSize = 2
	MaxSize = 3

	ChannelBufferSize = 1 << 18

	// MaxCounter = 695      // 4x4
	// MaxCounter = 124886 // 5x5
	MaxCounter = 88914655 // 6x6
)

type Canonicalizer struct {
	board *Board
	memo  *Memo
	key   *MemoKey
	moves [][]Move
}

func NewCanonicalizer() *Canonicalizer {
	moves := make([][]Move, 4096)
	return &Canonicalizer{moves: moves}
}

func (c *Canonicalizer) IsCanonical(board *Board) bool {
	key := *board.MemoKey()
	c.board = board
	c.memo = NewMemo()
	c.key = &key
	return c.isCanonical(0, -1)
}

func (c *Canonicalizer) isCanonical(depth, previousPiece int) bool {
	board := c.board
	if board.MemoKey().Less(c.key, false) {
		return false
	}
	if !c.memo.Add(board.MemoKey(), 0) {
		return true
	}
	buf := &c.moves[depth]
	*buf = board.Moves(*buf)
	for _, move := range *buf {
		if move.Piece == 0 {
			continue
		}
		if move.Piece == previousPiece {
			continue
		}
		board.DoMove(move)
		ok := c.isCanonical(depth+1, move.Piece)
		board.UndoMove(move)
		if !ok {
			return false
		}
	}
	return true
}

type Result struct {
	Board           *Board
	Unsolved        *Board
	Solution        Solution
	Group           int
	Counter         uint64
	JobCount        int
	CanonicalCount  int
	NonTrivialCount int
	MinimalCount    int
	Done            bool
}

func worker(jobs <-chan EnumeratorItem, results chan<- Result) {
	canonicalizer := NewCanonicalizer()
	var (
		jobCount        int
		canonicalCount  int
		nonTrivialCount int
		minimalCount    int
	)
	for job := range jobs {
		jobCount++

		board := job.Board

		// only evaluate "canonical" boards
		board.SortPieces()
		if !canonicalizer.IsCanonical(board) {
			continue
		}
		canonicalCount++

		// "unsolve" to find hardest reachable position
		unsolver := NewUnsolverWithStaticAnalyzer(board, nil)
		unsolved, solution := unsolver.UnsafeUnsolve()
		unsolved.SortPieces()

		// only interested in "non-trivial" puzzles
		if solution.NumMoves < 2 {
			continue
		}
		nonTrivialCount++

		// if removing any piece does not affect the solution, skip
		ok := true
		for i := 1; i < len(unsolved.Pieces); i++ {
			b := unsolved.Copy()
			b.RemovePiece(i)
			s := b.UnsafeSolve()
			if s.NumMoves == solution.NumMoves && s.NumSteps == solution.NumSteps {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		minimalCount++

		// we are interested in this puzzle
		results <- Result{
			board, unsolved, solution, job.Group, job.Counter,
			jobCount, canonicalCount, nonTrivialCount, minimalCount, false}

		// reset deltas
		jobCount = 0
		canonicalCount = 0
		nonTrivialCount = 0
		minimalCount = 0
	}
	results <- Result{Done: true}
}

func main() {
	e := NewEnumerator(W, H, PrimaryRow, PrimarySize, MinSize, MaxSize)
	// fmt.Println(e.Count())
	// return
	jobs := e.Enumerate(ChannelBufferSize)
	results := make(chan Result, ChannelBufferSize)

	wn := runtime.NumCPU()
	for i := 0; i < wn; i++ {
		go worker(jobs, results)
	}

	seen := make(map[string]bool)
	groups := make(map[int]int)
	var (
		jobCount        int
		canonicalCount  int
		nonTrivialCount int
		minimalCount    int
	)
	start := time.Now()
	for result := range results {
		if result.Done {
			wn--
			if wn == 0 {
				break
			}
			continue
		}

		jobCount += result.JobCount
		canonicalCount += result.CanonicalCount
		nonTrivialCount += result.NonTrivialCount
		minimalCount += result.MinimalCount

		unsolved := result.Unsolved
		solution := result.Solution
		key := unsolved.Hash()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = true
		groups[result.Group]++

		pct := float64(result.Counter) / MaxCounter
		elapsed := time.Since(start)
		fmt.Printf(
			"%02d %02d %02d %s %d\n",
			solution.NumMoves, solution.NumSteps, len(unsolved.Pieces),
			key, solution.MemoSize)
		fmt.Fprintf(
			os.Stderr, "[%.9f] %d jobs, %d canonical, %d non-trivial, %d minimal, %d distinct, %d groups - %s\n",
			pct, jobCount, canonicalCount, nonTrivialCount, minimalCount,
			len(seen), len(groups), elapsed)
	}
	// 4x4 = 31 31 31
	// 5x5 = 2329 6073 2186
	// 702 1280 665
}
