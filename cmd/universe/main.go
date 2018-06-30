package main

import (
	"fmt"
	"sort"

	. "github.com/fogleman/rush"
)

/*

......

....BB
...BB.
..BB..
.BB...
BB....

...BBB
..BBB.
.BBB..
BBB...

..BBCC
.BB.CC
.BBCC.
BB..CC
BB.CC.
BBCC..

.BBBCC
BBB.CC
BBBCC.

.BBCCC
BB.CCC
BBCCC.

6 groups: [[], [2], [3], [2, 2], [3, 2], [2, 3]]
1 primary row
5 rows
6 cols

6^11 = 362,797,056

1. precompute all possible rows/cols with bit masks
2. recursively pick rows/cols, check & mask
3. check if canonical (any position is Less - including primary!)
4. => worker unsolve+less => result
5. => check memo => write result

1. position generator routine
2. worker routines (unsolvers)
3. result handler routine

Rules:

A row cannot be completely filled with horizontal pieces.

A column cannot be completely filled with vertical pieces.

The primary row can only have one horizontal piece: the primary piece itself.

The primary piece cannot start in the winning position.

*/

type positionEntry struct {
	Pieces []Piece
	Mask   uint64
	Group  int
}

func makePositionEntry(w int, pieces []Piece, groups [][]int) positionEntry {
	ps := make([]Piece, len(pieces))
	copy(ps, pieces)
	var mask uint64
	for _, piece := range ps {
		idx := piece.Position
		stride := piece.Stride(w)
		for i := 0; i < piece.Size; i++ {
			mask |= 1 << uint(idx)
			idx += stride
		}
	}
	group := -1
	for i, g := range groups {
		if len(g) != len(pieces) {
			continue
		}
		ok := true
		for j := range g {
			if g[j] != pieces[j].Size {
				ok = false
			}
		}
		if ok {
			group = i
			break
		}
	}
	if group < 0 {
		panic("no group match")
	}
	return positionEntry{ps, mask, group}
}

type PositionGenerator struct {
	Width           int
	Height          int
	PrimaryRow      int
	PrimarySize     int
	MinSize         int
	MaxSize         int
	groups          [][]int
	rowEntries      [][]positionEntry
	colEntries      [][]positionEntry
	hardestBoard    *Board
	hardestSolution Solution
	counter1        uint64
	counter2        uint64
	prevGroup       int
}

func NewPositionGenerator(w, h, pr, ps, mins, maxs int) *PositionGenerator {
	pg := PositionGenerator{}
	pg.Width = w
	pg.Height = h
	pg.PrimaryRow = pr
	pg.PrimarySize = ps
	pg.MinSize = mins
	pg.MaxSize = maxs
	pg.rowEntries = make([][]positionEntry, h)
	pg.colEntries = make([][]positionEntry, w)
	pg.precomputeGroups(nil, 0)
	pg.precomputePositionEntries()
	return &pg
}

func NewDefaultPositionGenerator() *PositionGenerator {
	return NewPositionGenerator(5, 5, 2, 2, 2, 3)
}

func (pg *PositionGenerator) precomputeGroups(sizes []int, sum int) {
	if sum >= pg.Width {
		return
	}

	sizesCopy := make([]int, len(sizes))
	copy(sizesCopy, sizes)
	pg.groups = append(pg.groups, sizesCopy)

	n := len(sizes)
	for s := pg.MinSize; s <= pg.MaxSize; s++ {
		sizes = append(sizes, s)
		pg.precomputeGroups(sizes, sum+s)
		sizes = sizes[:n]
	}
}

func (pg *PositionGenerator) precomputeRow(y, x int, pieces []Piece) {
	w := pg.Width
	if x >= w {
		if y == pg.PrimaryRow {
			if len(pieces) != 1 {
				return
			}
			if pieces[0].Size != pg.PrimarySize {
				return
			}
			piece := pieces[0]
			target := (piece.Row(w)+1)*w - piece.Size
			if piece.Position == target {
				return
			}
		}
		var n int
		for _, piece := range pieces {
			n += piece.Size
		}
		if n >= w {
			return
		}
		pe := makePositionEntry(w, pieces, pg.groups)
		pg.rowEntries[y] = append(pg.rowEntries[y], pe)
		return
	}
	for s := pg.MinSize; s <= pg.MaxSize; s++ {
		if x+s > w {
			continue
		}
		p := y*w + x
		pieces = append(pieces, Piece{p, s, Horizontal})
		pg.precomputeRow(y, x+s, pieces)
		pieces = pieces[:len(pieces)-1]
	}
	pg.precomputeRow(y, x+1, pieces)
}

func (pg *PositionGenerator) precomputeCol(x, y int, pieces []Piece) {
	w := pg.Width
	h := pg.Height
	if y >= h {
		var n int
		for _, piece := range pieces {
			n += piece.Size
		}
		if n >= h {
			return
		}
		pe := makePositionEntry(w, pieces, pg.groups)
		pg.colEntries[x] = append(pg.colEntries[x], pe)
		return
	}
	for s := pg.MinSize; s <= pg.MaxSize; s++ {
		if y+s > h {
			continue
		}
		p := y*w + x
		pieces = append(pieces, Piece{p, s, Vertical})
		pg.precomputeCol(x, y+s, pieces)
		pieces = pieces[:len(pieces)-1]
	}
	pg.precomputeCol(x, y+1, pieces)
}

func (pg *PositionGenerator) precomputePositionEntries() {
	for y := 0; y < pg.Height; y++ {
		pg.precomputeRow(y, 0, nil)
	}
	for x := 0; x < pg.Width; x++ {
		pg.precomputeCol(x, 0, nil)
	}

	for y := 0; y < pg.Height; y++ {
		a := pg.rowEntries[y]
		sort.SliceStable(a, func(i, j int) bool { return a[i].Group < a[j].Group })
	}
	for x := 0; x < pg.Width; x++ {
		a := pg.colEntries[x]
		sort.SliceStable(a, func(i, j int) bool { return a[i].Group < a[j].Group })
	}
}

func (pg *PositionGenerator) populatePrimary() {
	var mask uint64
	board := NewEmptyBoard(pg.Width, pg.Height)
	for _, pe := range pg.rowEntries[pg.PrimaryRow] {
		mask |= pe.Mask
		for _, piece := range pe.Pieces {
			board.AddPiece(piece)
		}
		pg.populateRow(0, mask, 0, board)
		for range pe.Pieces {
			board.RemoveLastPiece()
		}
		mask ^= pe.Mask
	}
}

func (pg *PositionGenerator) populateRow(y int, mask uint64, group int, board *Board) {
	if y >= pg.Height {
		pg.populateCol(0, mask, group, board)
		return
	}
	if y == pg.PrimaryRow {
		pg.populateRow(y+1, mask, group, board)
		return
	}
	group *= len(pg.groups)
	for _, pe := range pg.rowEntries[y] {
		if mask&pe.Mask != 0 {
			continue
		}
		mask |= pe.Mask
		for _, piece := range pe.Pieces {
			board.AddPiece(piece)
		}
		pg.populateRow(y+1, mask, group+pe.Group, board)
		for range pe.Pieces {
			board.RemoveLastPiece()
		}
		mask ^= pe.Mask
	}
}

func (pg *PositionGenerator) populateCol(x int, mask uint64, group int, board *Board) {
	if x >= pg.Width {
		pg.counter1++
		if !pg.isCanonical(board) {
			return
		}
		// pg.hardest(board)
		// if !pg.hardestSolution.Solvable {
		// 	return
		// }
		pg.counter2++
		// if pg.counter1%1000000000 == 0 {
		// 	fmt.Println(pg.counter1, pg.counter2, group)
		// }
		// hardest := pg.hardestBoard
		// solution := pg.hardestSolution
		// eq := ""
		// if group == pg.prevGroup {
		// 	eq = "***"
		// }
		// pg.prevGroup = group
		// fmt.Println(pg.counter1, pg.counter2, hardest.Hash(), solution.NumMoves, solution.MemoSize, group, eq)
		// fmt.Println(pg.counter1, pg.counter2, board.Hash(), group, eq)
		return
	}
	group *= len(pg.groups)
	for _, pe := range pg.colEntries[x] {
		if mask&pe.Mask != 0 {
			continue
		}
		mask |= pe.Mask
		for _, piece := range pe.Pieces {
			board.AddPiece(piece)
		}
		pg.populateCol(x+1, mask, group+pe.Group, board)
		for range pe.Pieces {
			board.RemoveLastPiece()
		}
		mask ^= pe.Mask
	}
}

func (pg *PositionGenerator) canonicalSearch(board *Board, memo *Memo, key *MemoKey, previousPiece int) bool {
	if board.MemoKey().Less(key, true) {
		return false
	}
	if !memo.Add(board.MemoKey(), 0) {
		return true
	}
	for _, move := range board.Moves(nil) {
		if move.Piece == previousPiece {
			continue
		}
		board.DoMove(move)
		ok := pg.canonicalSearch(board, memo, key, move.Piece)
		board.UndoMove(move)
		if !ok {
			return false
		}
	}
	return true
}

func (pg *PositionGenerator) isCanonical(board *Board) bool {
	memo := NewMemo()
	key := *board.MemoKey()
	return pg.canonicalSearch(board, memo, &key, -1)
}

func (pg *PositionGenerator) hardestSearch(board *Board, memo *Memo, solver *Solver, previousPiece int) {
	if !memo.Add(board.MemoKey(), 0) {
		return
	}

	solution := solver.UnsafeSolve()
	delta := solution.NumMoves - pg.hardestSolution.NumMoves
	if delta > 0 || (delta == 0 && board.MemoKey().Less(pg.hardestBoard.MemoKey(), true)) {
		pg.hardestSolution = solution
		pg.hardestBoard = board.Copy()
	}

	for _, move := range board.Moves(nil) {
		if move.Piece == previousPiece {
			continue
		}
		board.DoMove(move)
		pg.hardestSearch(board, memo, solver, move.Piece)
		board.UndoMove(move)
	}
}

func (pg *PositionGenerator) hardest(board *Board) {
	memo := NewMemo()
	solver := NewSolver(board)
	pg.hardestBoard = board.Copy()
	pg.hardestSolution = solver.Solve()
	if !pg.hardestSolution.Solvable {
		return
	}
	pg.hardestSearch(board, memo, solver, -1)
	pg.hardestBoard.SortPieces()
}

func (pg *PositionGenerator) Generate() {
	pg.populatePrimary()
}

func main() {
	pg := NewDefaultPositionGenerator()
	pg.Generate()
	fmt.Println(pg.counter1, pg.counter2)
	// 27,103,652,326
	// 22,138,497,189
}
