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
	// MaxCounter = 124886   // 5x5
	MaxCounter = 88914655 // 6x6
)

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
		if !IsCanonical(board) {
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
	results <- Result{
		Done:            true,
		JobCount:        jobCount,
		CanonicalCount:  canonicalCount,
		NonTrivialCount: nonTrivialCount,
		MinimalCount:    minimalCount,
	}
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
		jobCount += result.JobCount
		canonicalCount += result.CanonicalCount
		nonTrivialCount += result.NonTrivialCount
		minimalCount += result.MinimalCount

		if result.Done {
			wn--
			if wn == 0 {
				break
			}
			continue
		}

		unsolved := result.Unsolved
		solution := result.Solution
		key := unsolved.Hash()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = true
		groups[result.Group]++

		pct := float64(result.Counter) / MaxCounter
		fmt.Printf(
			"%02d %02d %02d %s %d %d\n",
			solution.NumMoves, solution.NumSteps, len(unsolved.Pieces),
			key, solution.MemoSize, result.Group)
		fmt.Fprintf(
			os.Stderr, "[%.9f] %d in, %d cn, %d nt, %d mn, %d dt, %d gp - %s\n",
			pct, jobCount, canonicalCount, nonTrivialCount, minimalCount,
			len(seen), len(groups), time.Since(start))
	}

	fmt.Fprintf(
		os.Stderr, "[%.9f] %d in, %d cn, %d nt, %d mn, %d dt, %d gp - %s\n",
		1.0, jobCount, canonicalCount, nonTrivialCount, minimalCount,
		len(seen), len(groups), time.Since(start))
}
