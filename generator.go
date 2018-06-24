package rush

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
	board.AddPiece(Piece{g.PrimaryRow * g.Width, g.PrimarySize, Horizontal})

	// // hill climb
	// score := -board.Energy()
	// for i := 0; i < iterations; i++ {
	// 	undo := board.Mutate()
	// 	newScore := -board.Energy()
	// 	if newScore <= score {
	// 		undo()
	// 	} else {
	// 		score = newScore
	// 		fmt.Println(board)
	// 		fmt.Println(i, score)
	// 		fmt.Println()
	// 	}
	// }

	board = anneal(board, 20, 1, iterations)

	return board
}
