package rush

import (
	"math/rand"
)

// TODO: piece pool / bag - in use / out of use pieces

type Generator struct {
	Width       int
	Height      int
	PrimarySize int
	PrimaryRow  int
	// MinPieces   int
	// MaxPieces   int
	// MinSize     int
	// MaxSize     int
}

func NewDefaultGenerator() *Generator {
	return &Generator{6, 6, 2, 2}
}

func (g *Generator) Generate() *Board {
	// create empty board
	board := NewEmptyBoard(g.Width, g.Height)

	// place the primary piece
	primary := Piece{g.PrimaryRow * g.Width, g.PrimarySize, Horizontal}
	board.AddPiece(primary)

	// add random pieces
	n := rand.Intn(12) + 3
	for i := 0; i < n; i++ {
		piece, ok := g.randomPiece(board, 100)
		if ok {
			board.AddPiece(piece)
		}
	}
	return board
}

func (g *Generator) randomPiece(board *Board, maxAttempts int) (Piece, bool) {
	w := board.Width
	h := board.Height
	for i := 0; i < maxAttempts; i++ {
		size := 2 + rand.Intn(2) // TODO: weighted
		orientation := Orientation(rand.Intn(2))
		var x, y int
		if orientation == Vertical {
			x = rand.Intn(w)
			y = rand.Intn(h - size + 1)
		} else {
			x = rand.Intn(w - size + 1)
			y = rand.Intn(h)
		}
		position := y*w + x
		piece := Piece{position, size, orientation}
		if !board.isOccupied(piece) {
			return piece, true
		}
	}
	return Piece{}, false
}
