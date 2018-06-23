package rush

import (
	"fmt"
	"image"
	"sort"
	"strings"
)

type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

type Piece struct {
	Position    int
	Size        int
	Orientation Orientation
}

func (piece *Piece) Stride(w int) int {
	if piece.Orientation == Horizontal {
		return 1
	}
	return w
}

func (piece *Piece) Row(w int) int {
	return piece.Position / w
}

func (piece *Piece) Col(w int) int {
	return piece.Position % w
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

type Board struct {
	Width    int
	Height   int
	Pieces   []Piece
	Walls    []int
	occupied []bool
	memoKey  MemoKey
}

func NewEmptyBoard(w, h int) *Board {
	occupied := make([]bool, w*h)
	memoKey := MakeMemoKey(nil)
	return &Board{w, h, nil, nil, occupied, memoKey}
}

func NewBoard(desc []string) (*Board, error) {
	// determine board size
	h := len(desc)
	if h < MinBoardSize {
		return nil, fmt.Errorf("board height must be >= %d", MinBoardSize)
	}
	w := len(desc[0])
	if w < MinBoardSize {
		return nil, fmt.Errorf("board width must be >= %d", MinBoardSize)
	}

	// identify occupied cells and their labels
	occupied := make([]bool, w*h)
	positions := make(map[string][]int)
	var walls []int
	for y, row := range desc {
		for x, value := range row {
			label := string(value)
			if label == "." {
				continue
			}
			i := y*w + x
			occupied[i] = true
			if label == "x" {
				walls = append(walls, i)
			} else {
				positions[label] = append(positions[label], i)
			}

		}
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
		if len(ps) < MinPieceSize {
			return nil, fmt.Errorf("piece %s length must be >= %d", label, MinPieceSize)
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
	board := &Board{w, h, pieces, walls, occupied, MakeMemoKey(pieces)}
	return board, board.Validate()
}

func (board *Board) String() string {
	w := board.Width
	h := board.Height
	grid := make([]string, w*h)
	for i := range grid {
		grid[i] = "."
	}
	for _, i := range board.Walls {
		grid[i] = "x"
	}
	for i, piece := range board.Pieces {
		label := string('A' + i)
		idx := piece.Position
		stride := piece.Stride(w)
		for j := 0; j < piece.Size; j++ {
			grid[idx] = label
			idx += stride
		}
	}
	rows := make([]string, h)
	for y := 0; y < h; y++ {
		i := y * w
		rows[y] = strings.Join(grid[i:i+w], "")
	}
	return strings.Join(rows, "\n")
}

func (board *Board) Validate() error {
	w := board.Width
	h := board.Height
	pieces := board.Pieces

	// board size must be >= MinBoardSize
	if w < MinBoardSize {
		return fmt.Errorf("board width must be >= %d", MinBoardSize)
	}
	if h < MinBoardSize {
		return fmt.Errorf("board height must be >= %d", MinBoardSize)
	}

	// board must have at least one piece
	if len(pieces) < 1 {
		return fmt.Errorf("board must have at least one piece")
	}

	// board must have <= MaxPieces
	if len(pieces) > MaxPieces {
		return fmt.Errorf("board must have <= %d pieces", MaxPieces)
	}

	// primary piece must be horizontal
	if pieces[0].Orientation != Horizontal {
		return fmt.Errorf("primary piece must be horizontal")
	}

	// validate walls
	occupied := make([]bool, w*h)
	for _, i := range board.Walls {
		// wall must be inside the grid
		if i < 0 || i >= w*h {
			return fmt.Errorf("a wall is outside of the grid")
		}

		// walls must not intersect
		if occupied[i] {
			return fmt.Errorf("a wall intersects another wall")
		}
		occupied[i] = true
	}

	// validate pieces
	primaryRow := pieces[0].Row(w)
	for i, piece := range pieces {
		label := string('A' + i)
		row := piece.Row(w)
		col := piece.Col(w)

		// piece size must be >= MinPieceSize
		if piece.Size < MinPieceSize {
			return fmt.Errorf("piece %s must have size >= %d", label, MinPieceSize)
		}

		// no horizontal pieces can be on the same row as the primary piece
		if i > 0 && piece.Orientation == Horizontal && row == primaryRow {
			return fmt.Errorf("no horizontal pieces can be on the primary row")
		}

		// pieces must be contained within the grid
		if piece.Orientation == Horizontal {
			if row < 0 || row >= h || col < 0 || col+piece.Size > w {
				return fmt.Errorf("piece %s is outside of the grid", label)
			}
		} else {
			if col < 0 || col >= w || row < 0 || row+piece.Size > h {
				return fmt.Errorf("piece %s is outside of the grid", label)
			}
		}

		// pieces must not intersect
		idx := piece.Position
		stride := piece.Stride(w)
		for j := 0; j < piece.Size; j++ {
			if occupied[idx] {
				return fmt.Errorf("piece %s intersects with another piece", label)
			}
			occupied[idx] = true
			idx += stride
		}
	}

	return nil
}

func (board *Board) isOccupied(piece Piece) bool {
	idx := piece.Position
	stride := piece.Stride(board.Width)
	for i := 0; i < piece.Size; i++ {
		if board.occupied[idx] {
			return true
		}
		idx += stride
	}
	return false
}

func (board *Board) setOccupied(piece Piece, value bool) {
	idx := piece.Position
	stride := piece.Stride(board.Width)
	for i := 0; i < piece.Size; i++ {
		board.occupied[idx] = value
		idx += stride
	}
}

func (board *Board) AddPiece(piece Piece) bool {
	if board.isOccupied(piece) {
		return false
	}
	i := len(board.Pieces)
	board.Pieces = append(board.Pieces, piece)
	board.setOccupied(piece, true)
	board.memoKey[i] = piece.Position
	return true
}

func (board *Board) AddWall(i int) bool {
	if board.occupied[i] {
		return false
	}
	board.Walls = append(board.Walls, i)
	board.occupied[i] = true
	return true
}

func (board *Board) Target() int {
	w := board.Width
	piece := board.Pieces[0]
	row := piece.Row(w)
	return (row+1)*w - piece.Size
}

func (board *Board) Moves(buf []Move) []Move {
	moves := buf[:0]
	w := board.Width
	h := board.Height
	for i, piece := range board.Pieces {
		var stride, reverseSteps, forwardSteps int
		if piece.Orientation == Vertical {
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
			if board.occupied[idx] {
				break
			}
			moves = append(moves, Move{i, steps})
			idx -= stride
		}
		// forward (positive steps)
		idx = piece.Position + piece.Size*stride
		for steps := 1; steps <= forwardSteps; steps++ {
			if board.occupied[idx] {
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
	stride := piece.Stride(board.Width)

	idx := piece.Position
	for i := 0; i < piece.Size; i++ {
		board.occupied[idx] = false
		idx += stride
	}

	piece.Position += stride * move.Steps
	board.memoKey[move.Piece] = piece.Position

	idx = piece.Position
	for i := 0; i < piece.Size; i++ {
		board.occupied[idx] = true
		idx += stride
	}
}

func (board *Board) UndoMove(move Move) {
	board.DoMove(Move{move.Piece, -move.Steps})
}

func (board *Board) MemoKey() *MemoKey {
	return &board.memoKey
}

func (board *Board) Solve() Solution {
	return NewSolver(board).Solve()
}

func (board *Board) Render() image.Image {
	return renderBoard(board)
}

func (board *Board) Impossible() bool {
	return theStaticAnalyzer.Impossible(board)
}

func (board *Board) BlockedSquares() []int {
	return theStaticAnalyzer.BlockedSquares(board)
}
