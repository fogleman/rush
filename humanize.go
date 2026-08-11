package rush

import (
	"fmt"
	"sort"
)

// moveGraph captures the ordering constraints between the moves of a solution.
// Two moves commute if and only if the cells they sweep are disjoint - if the
// swept cells touch, one move has to vacate or pass through a cell the other
// one needs, which forces an order between them. Everything else is
// unconstrained, so the resequencings of a solution that still solve the board
// are exactly the topological orderings of this graph.
//
// The same-piece case falls out for free: a piece's second move starts where
// its first move ended, so their swept cells always overlap.
type moveGraph struct {
	moves []Move

	// prereqs[i] lists the moves that must happen before moves[i], excluding
	// any that are already implied by another prerequisite, ordered so that
	// the longest chain of setup work comes first.
	prereqs [][]int
}

func newMoveGraph(board *Board, moves []Move) *moveGraph {
	n := len(moves)

	// walk the sequence, recording the cells swept by each move - every cell
	// its piece occupies at any point during the slide, including both the
	// starting and ending footprints - plus where the piece sat beforehand
	b := board.Copy()
	swept := make([][]bool, n)
	positions := make([]int, n)
	for i, move := range moves {
		piece := b.Pieces[move.Piece]
		stride := piece.Stride(b.Width)
		lo := piece.Position
		hi := lo + (piece.Size-1)*stride
		if move.Steps < 0 {
			lo += move.Steps * stride
		} else {
			hi += move.Steps * stride
		}
		cells := make([]bool, len(b.occupied))
		for idx := lo; idx <= hi; idx += stride {
			cells[idx] = true
		}
		swept[i] = cells
		positions[i] = piece.Position
		b.DoMove(move)
	}

	// an earlier move constrains a later one if their swept cells overlap
	succs := make([][]int, n)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if overlaps(swept[i], swept[j]) {
				succs[i] = append(succs[i], j)
			}
		}
	}

	// reachable[i][j] means move j transitively depends on move i. The moves
	// arrive in a valid order already, so one reverse pass computes it.
	reachable := make([][]bool, n)
	for i := n - 1; i >= 0; i-- {
		r := make([]bool, n)
		for _, j := range succs[i] {
			r[j] = true
			for k, ok := range reachable[j] {
				if ok {
					r[k] = true
				}
			}
		}
		reachable[i] = r
	}

	// how much has to happen before each move, used below to order siblings
	depth := make([]int, n)
	for _, r := range reachable {
		for j, ok := range r {
			if ok {
				depth[j]++
			}
		}
	}

	// keep only the direct prerequisites, dropping i -> j whenever i already
	// reaches j the long way around
	prereqs := make([][]int, n)
	for i, s := range succs {
		for _, j := range s {
			direct := true
			for _, k := range s {
				if k != j && reachable[k][j] {
					direct = false
					break
				}
			}
			if direct {
				prereqs[j] = append(prereqs[j], i)
			}
		}
	}

	// Order each move's prerequisites for the walk in humanize. Deepest first
	// puts the long chain of accumulated setup work at the front, which leaves
	// the one-off "just get this car out of the way" moves sitting right next
	// to the move that needed them. Genuine ties are arbitrary, so break them
	// in row-major order: top to bottom, left to right.
	for _, p := range prereqs {
		sort.Slice(p, func(a, b int) bool {
			x, y := p[a], p[b]
			if depth[x] != depth[y] {
				return depth[x] > depth[y]
			}
			if positions[x] != positions[y] {
				return positions[x] < positions[y]
			}
			return x < y
		})
	}

	return &moveGraph{moves, prereqs}
}

func overlaps(a, b []bool) bool {
	for i, ok := range a {
		if ok && b[i] {
			return true
		}
	}
	return false
}

// A move that sits more than maxSpan moves away from the move it enables has
// already been forgotten by the time it pays off, and pulling it from twelve
// moves away to eleven doesn't help anyone follow along. Charging no more than
// maxSpan for a gap keeps polish from trading a tidy local ordering for a
// token gain on a dependency that stays out of sight either way.
const maxSpan = 4

// stretch measures how scattered an ordering is: for each move, how far it
// sits from the moves that directly depend on it, counting anything past
// maxSpan as simply far away. Clearing a car out of the way and immediately
// using the space it freed scores lower than doing unrelated things in
// between. Lower is more focused.
func (g *moveGraph) stretch(order, position []int) int {
	for i, j := range order {
		position[j] = i
	}
	var total int
	for i, prereqs := range g.prereqs {
		for _, j := range prereqs {
			span := position[i] - position[j]
			if span > maxSpan {
				span = maxSpan
			}
			total += span
		}
	}
	return total
}

// polish slides moves around until no single move wants to be anywhere else.
// Every relocation strictly lowers stretch, which is what makes it terminate.
//
// This is what keeps a move that several later moves are waiting on from being
// stranded next to whichever one happened to claim it first.
func (g *moveGraph) polish(order []int) {
	position := make([]int, len(g.moves))
	trial := make([]int, len(order))
	for improved := true; improved; {
		improved = false
		for i := range order {
			if g.shift(order, i, trial, position) {
				improved = true
			}
		}
	}
}

// shift relocates the move at position p to wherever it lowers stretch most,
// and reports whether it moved anything. Ties leave the move where it is, so
// the ordering walk chose survives anywhere the choice doesn't matter.
func (g *moveGraph) shift(order []int, p int, trial, position []int) bool {
	x := order[p]

	// How far it can travel in either direction before running into a move it
	// has to stay on one side of. A move only ever steps past its immediate
	// neighbor, so having no direct dependency between the two is the whole
	// test - anything related indirectly has a third move in between.
	lo := p
	for lo > 0 && !g.isPrereq(order[lo-1], x) {
		lo--
	}
	hi := p
	for hi < len(order)-1 && !g.isPrereq(x, order[hi+1]) {
		hi++
	}

	best, target := g.stretch(order, position), p
	for q := lo; q <= hi; q++ {
		if q == p {
			continue
		}
		copy(trial, order)
		relocate(trial, p, q)
		if s := g.stretch(trial, position); s < best {
			best, target = s, q
		}
	}

	if target == p {
		return false
	}
	relocate(order, p, target)
	return true
}

// relocate moves the element at index p to index q, sliding whatever is in
// between over to make room.
func relocate(a []int, p, q int) {
	x := a[p]
	if q > p {
		copy(a[p:q], a[p+1:q+1])
	} else {
		copy(a[q+1:p+1], a[q:p])
	}
	a[q] = x
}

// isPrereq reports whether move i has to happen before move j. Only correct
// for moves that are adjacent in a valid ordering, which is all polish needs:
// anything reaching j indirectly would have to sit between them.
func (g *moveGraph) isPrereq(i, j int) bool {
	for _, k := range g.prereqs[j] {
		if k == i {
			return true
		}
	}
	return false
}

// humanize reorders the moves of a solution so that it is easier to follow. A
// solver is free to interleave moves that serve unrelated purposes, since it
// only cares about reaching the goal in as few moves as possible; a person
// works one subgoal at a time. The reordering never adds, drops or alters a
// move, so the result solves the board in exactly as many moves as the input.
func humanize(board *Board, moves []Move) []Move {
	result := make([]Move, 0, len(moves))
	if len(moves) == 0 {
		return result
	}

	g := newMoveGraph(board, moves)
	order := g.walk()
	g.polish(order)
	for _, i := range order {
		result = append(result, moves[i])
	}
	return result
}

// walk orders the moves by working backwards from the winning move, emitting
// everything a move depends on immediately before the move itself. Setup work
// ends up adjacent to whatever it was setting up, recursively, so the sequence
// reads as a series of subgoals - clear this car, which needs that one moved
// first, now advance - instead of a shuffle of unrelated moves.
func (g *moveGraph) walk() []int {
	order := make([]int, 0, len(g.moves))
	done := make([]bool, len(g.moves))

	var emit func(i int)
	emit = func(i int) {
		if done[i] {
			return
		}
		for _, j := range g.prereqs[i] {
			emit(j)
		}
		done[i] = true
		order = append(order, i)
	}

	emit(len(g.moves) - 1)

	// An optimal solution has no move the winning move doesn't depend on, or
	// it could be dropped for a shorter solution, but don't lose any moves if
	// we were handed something other than an optimal solution.
	for i := range g.moves {
		emit(i)
	}

	return order
}

// moveStretch scores how scattered a move sequence is. See moveGraph.stretch.
func moveStretch(board *Board, moves []Move) int {
	g := newMoveGraph(board, moves)
	order := make([]int, len(moves))
	for i := range order {
		order[i] = i
	}
	return g.stretch(order, make([]int, len(moves)))
}

// validateSolution reports whether every move is legal from the board's
// current position and whether the sequence leaves the primary piece at its
// target.
func validateSolution(board *Board, moves []Move) error {
	b := board.Copy()
	var buf []Move
	for i, move := range moves {
		buf = b.Moves(buf)
		var legal bool
		for _, m := range buf {
			if m == move {
				legal = true
				break
			}
		}
		if !legal {
			return fmt.Errorf("move %d (%v) is not legal", i+1, move)
		}
		b.DoMove(move)
	}
	if b.Pieces[0].Position != b.Target() {
		return fmt.Errorf("moves do not solve the board")
	}
	return nil
}
