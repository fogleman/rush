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

func (g *Generator) Generate(numPieces, numWalls int) *Board {
	// create empty board
	board := NewEmptyBoard(g.Width, g.Height)

	// place the primary piece
	primary := Piece{g.PrimaryRow * g.Width, g.PrimarySize, Horizontal}
	board.AddPiece(primary)

	// add random pieces
	for i := 0; i < numPieces; i++ {
		if piece, ok := board.randomPiece(100); ok {
			board.AddPiece(piece)
		}
	}

	// add random walls
	for i := 0; i < numWalls; i++ {
		if wall, ok := board.randomWall(100); ok {
			board.AddWall(wall)
		}
	}
	return board
}
