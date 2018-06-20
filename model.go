package rush

import (
	"fmt"
	"sort"
	"strings"
)

type Direction int

const (
	Horizontal Direction = iota
	Vertical
)

type Piece struct {
	Position  int
	Size      int
	Direction Direction
}

type Board struct {
	Width    int
	Height   int
	Pieces   []Piece
	Occupied []bool
}

type Move struct {
	Piece int
	Steps int
}

func (move Move) AbsSteps() int {
	if move.Steps < 0 {
		return -move.Steps
	}
	return move.Steps
}

func NewBoard(desc []string) (*Board, error) {
	// determine board size
	h := len(desc)
	if h < 2 {
		return nil, fmt.Errorf("board height must be >= 2")
	}
	w := len(desc[0])
	if w < 2 {
		return nil, fmt.Errorf("board width must be >= 2")
	}

	// identify occupied cells and their labels
	occupied := make([]bool, w*h)
	positions := make(map[string][]int)
	for y, row := range desc {
		for x, value := range row {
			label := string(value)
			if label == "." {
				continue
			}
			i := y*w + x
			occupied[i] = true
			positions[label] = append(positions[label], i)
		}
	}
	if len(positions) < 1 {
		return nil, fmt.Errorf("board must have at least one piece")
	}

	// find and sort distinct piece labels
	labels := make([]string, 0, len(positions))
	for label := range positions {
		labels = append(labels, label)
	}
	sort.Strings(labels)

	// validate and create pieces
	pieces := make([]Piece, 0, len(labels))
	for _, label := range labels {
		ps := positions[label]
		if len(ps) < 2 {
			return nil, fmt.Errorf("piece %s length must be >= 2", label)
		}
		stride := ps[1] - ps[0]
		if stride != 1 && stride != w {
			return nil, fmt.Errorf("piece %s has invalid shape", label)
		}
		for i := 2; i < len(ps); i++ {
			if ps[i]-ps[i-1] != stride {
				return nil, fmt.Errorf("piece %s has invalid shape", label)
			}
		}
		dir := Horizontal
		if stride != 1 {
			dir = Vertical
		}
		pieces = append(pieces, Piece{ps[0], len(ps), dir})
	}

	// create board
	return &Board{w, h, pieces, occupied}, nil
}

func (board *Board) String() string {
	w := board.Width
	h := board.Height
	grid := make([]string, w*h)
	for i := range grid {
		grid[i] = "."
	}
	for i, piece := range board.Pieces {
		label := string('A' + i)
		stride := 1
		if piece.Direction == Vertical {
			stride = w
		}
		for j := 0; j < piece.Size; j++ {
			grid[piece.Position+stride*j] = label
		}
	}
	rows := make([]string, h)
	for y := 0; y < h; y++ {
		rows[y] = strings.Join(grid[y*w:y*w+w], "")
	}
	return strings.Join(rows, "\n")
}

func (board *Board) Moves(buf []Move) []Move {
	moves := buf[:0]
	w := board.Width
	h := board.Height
	for i, piece := range board.Pieces {
		var stride, reverseSteps, forwardSteps int
		if piece.Direction == Vertical {
			y := piece.Position / w
			reverseSteps = -y
			forwardSteps = h - piece.Size - y
			stride = w
		} else {
			x := piece.Position % w
			reverseSteps = -x
			forwardSteps = w - piece.Size - x
			stride = 1
		}
		// reverse (negative steps)
		idx := piece.Position - stride
		for steps := -1; steps >= reverseSteps; steps-- {
			if board.Occupied[idx] {
				break
			}
			moves = append(moves, Move{i, steps})
			idx -= stride
		}
		// forward (positive steps)
		idx = piece.Position + piece.Size*stride
		for steps := 1; steps <= forwardSteps; steps++ {
			if board.Occupied[idx] {
				break
			}
			moves = append(moves, Move{i, steps})
			idx += stride
		}
	}
	return moves
}

func (board *Board) DoMove(move Move) {
	piece := &board.Pieces[move.Piece]
	stride := 1
	if piece.Direction == Vertical {
		stride = board.Width
	}
	idx := piece.Position
	for i := 0; i < piece.Size; i++ {
		board.Occupied[idx] = false
		idx += stride
	}
	piece.Position += stride * move.Steps
	idx = piece.Position
	for i := 0; i < piece.Size; i++ {
		board.Occupied[idx] = true
		idx += stride
	}
}

func (board *Board) UndoMove(move Move) {
	board.DoMove(Move{move.Piece, -move.Steps})
}

func (board *Board) MemoKey() MemoKey {
	var key MemoKey
	for i, piece := range board.Pieces {
		key[i] = piece.Position
	}
	return key
}

func (board *Board) Solve(target int) []Move {
	solver := NewSolver(board, target)
	return solver.Solve()
}
