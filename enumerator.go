package rush

import (
	"math"
	"sort"
)

type positionEntry struct {
	Pieces  []Piece
	Mask    uint64
	Require uint64
	Group   int
}

func makePositionEntry(stride int, noRequire uint64, pieces []Piece, groups [][]int) positionEntry {
	ps := make([]Piece, len(pieces))
	copy(ps, pieces)
	var mask uint64
	for _, piece := range ps {
		idx := piece.Position
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
				break
			}
		}
		if ok {
			group = i
			break
		}
	}
	if group < 0 {
		panic("makePositionEntry failed")
	}
	require := (mask >> uint(stride)) & ^mask & ^noRequire
	return positionEntry{ps, mask, require, group}
}

type EnumeratorItem struct {
	Board   *Board
	Group   int
	Counter uint64
}

type Enumerator struct {
	width       int
	height      int
	primaryRow  int
	primarySize int
	minSize     int
	maxSize     int
	noRequire   uint64
	groups      [][]int
	rowEntries  [][]positionEntry
	colEntries  [][]positionEntry
}

func NewEnumerator(w, h, pr, ps, mins, maxs int) *Enumerator {
	e := Enumerator{}
	e.width = w
	e.height = h
	e.primaryRow = pr
	e.primarySize = ps
	e.minSize = mins
	e.maxSize = maxs
	e.rowEntries = make([][]positionEntry, h)
	e.colEntries = make([][]positionEntry, w)
	for y := 0; y <= h; y++ {
		e.noRequire |= 1 << uint(y*w+w-1)
	}
	e.precomputeGroups(nil, 0)
	e.precomputePositionEntries()
	return &e
}

func NewDefaultEnumerator() *Enumerator {
	return NewEnumerator(6, 6, 2, 2, 2, 3)
}

func (e *Enumerator) Enumerate(channelBufferSize int) <-chan EnumeratorItem {
	ch := make(chan EnumeratorItem, channelBufferSize)
	go func() {
		e.populatePrimaryRow(ch)
		close(ch)
	}()
	return ch
}

func (e *Enumerator) MaxGroup() int {
	n := e.width + e.height - 1
	return int(math.Pow(float64(len(e.groups)), float64(n)))
}

func (e *Enumerator) Count() uint64 {
	// 4x4 = 2896
	// 5x5 = 1566424
	// 6x6 = 4965155137

	// with require mask:
	// 4x4 = 695
	// 5x5 = 124886
	// 6x6 = 88914655
	return e.countPrimaryRow()
}

func (e *Enumerator) precomputeGroups(sizes []int, sum int) {
	if sum >= e.width {
		return
	}

	sizesCopy := make([]int, len(sizes))
	copy(sizesCopy, sizes)
	e.groups = append(e.groups, sizesCopy)

	n := len(sizes)
	for s := e.minSize; s <= e.maxSize; s++ {
		sizes = append(sizes, s)
		e.precomputeGroups(sizes, sum+s)
		sizes = sizes[:n]
	}
}

func (e *Enumerator) precomputeRow(y, x int, pieces []Piece) {
	w := e.width
	if x >= w {
		if y == e.primaryRow {
			if len(pieces) != 1 {
				return
			}
			if pieces[0].Size != e.primarySize {
				return
			}
			piece := pieces[0]
			target := (piece.Row(w)+1)*w - piece.Size
			if piece.Position != target {
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
		pe := makePositionEntry(1, e.noRequire, pieces, e.groups)
		e.rowEntries[y] = append(e.rowEntries[y], pe)
		return
	}
	for s := e.minSize; s <= e.maxSize; s++ {
		if x+s > w {
			continue
		}
		p := y*w + x
		pieces = append(pieces, Piece{p, s, Horizontal})
		e.precomputeRow(y, x+s, pieces)
		pieces = pieces[:len(pieces)-1]
	}
	e.precomputeRow(y, x+1, pieces)
}

func (e *Enumerator) precomputeCol(x, y int, pieces []Piece) {
	w := e.width
	h := e.height
	if y >= h {
		var n int
		for _, piece := range pieces {
			n += piece.Size
		}
		if n >= h {
			return
		}
		pe := makePositionEntry(w, 0, pieces, e.groups)
		e.colEntries[x] = append(e.colEntries[x], pe)
		return
	}
	for s := e.minSize; s <= e.maxSize; s++ {
		if y+s > h {
			continue
		}
		p := y*w + x
		pieces = append(pieces, Piece{p, s, Vertical})
		e.precomputeCol(x, y+s, pieces)
		pieces = pieces[:len(pieces)-1]
	}
	e.precomputeCol(x, y+1, pieces)
}

func (e *Enumerator) precomputePositionEntries() {
	for y := 0; y < e.height; y++ {
		e.precomputeRow(y, 0, nil)
	}
	for x := 0; x < e.width; x++ {
		e.precomputeCol(x, 0, nil)
	}

	for y := 0; y < e.height; y++ {
		a := e.rowEntries[y]
		sort.SliceStable(a, func(i, j int) bool { return a[i].Group < a[j].Group })
	}
	for x := 0; x < e.width; x++ {
		a := e.colEntries[x]
		sort.SliceStable(a, func(i, j int) bool { return a[i].Group < a[j].Group })
	}
}

func (e *Enumerator) populatePrimaryRow(ch chan EnumeratorItem) {
	var counter uint64
	board := NewEmptyBoard(e.width, e.height)
	for _, pe := range e.rowEntries[e.primaryRow] {
		for _, piece := range pe.Pieces {
			board.addPiece(piece)
		}
		e.populateRow(ch, &counter, 0, pe.Mask, 0, 0, board)
		for range pe.Pieces {
			board.RemoveLastPiece()
		}
	}
}

func (e *Enumerator) populateRow(ch chan EnumeratorItem, counter *uint64, y int, mask, require uint64, group int, board *Board) {
	if y >= e.height {
		e.populateCol(ch, counter, 0, mask, require, group, board)
		return
	}
	if y == e.primaryRow {
		e.populateRow(ch, counter, y+1, mask, require, group, board)
		return
	}
	group *= len(e.groups)
	for _, pe := range e.rowEntries[y] {
		if mask&pe.Mask != 0 {
			continue
		}
		for _, piece := range pe.Pieces {
			board.addPiece(piece)
		}
		e.populateRow(ch, counter, y+1, mask|pe.Mask, require|pe.Require, group+pe.Group, board)
		for range pe.Pieces {
			board.RemoveLastPiece()
		}
	}
}

func (e *Enumerator) populateCol(ch chan EnumeratorItem, counter *uint64, x int, mask, require uint64, group int, board *Board) {
	if x >= e.width {
		if mask&require != require {
			return
		}
		*counter++
		ch <- EnumeratorItem{board.Copy(), group, *counter}
		return
	}
	group *= len(e.groups)
	for _, pe := range e.colEntries[x] {
		if mask&pe.Mask != 0 {
			continue
		}
		for _, piece := range pe.Pieces {
			board.addPiece(piece)
		}
		e.populateCol(ch, counter, x+1, mask|pe.Mask, require|pe.Require, group+pe.Group, board)
		for range pe.Pieces {
			board.RemoveLastPiece()
		}
	}
}

func (e *Enumerator) countPrimaryRow() uint64 {
	var counter uint64
	for _, pe := range e.rowEntries[e.primaryRow] {
		e.countRow(0, pe.Mask, 0, &counter)
	}
	return counter
}

func (e *Enumerator) countRow(y int, mask, require uint64, counter *uint64) {
	if y >= e.height {
		e.countCol(0, mask, require, counter)
		return
	}
	if y == e.primaryRow {
		e.countRow(y+1, mask, require, counter)
		return
	}
	for _, pe := range e.rowEntries[y] {
		if mask&pe.Mask != 0 {
			continue
		}
		e.countRow(y+1, mask|pe.Mask, require|pe.Require, counter)
	}
}

func (e *Enumerator) countCol(x int, mask, require uint64, counter *uint64) {
	if x >= e.width {
		if mask&require != require {
			return
		}
		*counter++
		return
	}
	for _, pe := range e.colEntries[x] {
		if mask&pe.Mask != 0 {
			continue
		}
		e.countCol(x+1, mask|pe.Mask, require|pe.Require, counter)
	}
}
