package rush

import "fmt"

// TODO: piece pool / bag - in use / out of use pieces?

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
	primary := Piece{g.PrimaryRow * g.Width, g.PrimarySize, Horizontal}
	board.AddPiece(primary)

	// hill climb
	score := g.score(board)
	for i := 0; i < iterations; i++ {
		undo := board.Mutate()
		newScore := g.score(board)
		if newScore <= score {
			undo()
		} else {
			score = newScore
			fmt.Println(i, score)
		}
	}

	return board
}

func (g *Generator) score(board *Board) int {
	solution := board.Solve()
	if !solution.Solvable {
		return -1
	}
	return solution.NumMoves
}
