package rush

import "fmt"

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

func (g *Generator) Generate(iterations int) *Board {
	// create empty board
	board := NewEmptyBoard(g.Width, g.Height)

	// place the primary piece
	board.AddPiece(Piece{g.PrimaryRow * g.Width, g.PrimarySize, Horizontal})

	// simulated annealing
	board = anneal(board, 20, 0.5, iterations)

	// unsolve step
	before := NewSolver(board).Solve().NumMoves
	board, _ = NewUnsolver(board).Unsolve()
	after := NewSolver(board).Solve().NumMoves
	fmt.Println(before, after)

	return board
}
