package rush

import (
	"fmt"
	"sort"
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
	Width  int
	Height int
	// Target int
	Pieces   []Piece
	Occupied []bool
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
			c := string(value)
			if c == "." {
				continue
			}
			i := y*w + x
			occupied[i] = true
			positions[c] = append(positions[c], i)
		}
	}
	if len(positions) < 1 {
		return nil, fmt.Errorf("board must have at least one piece")
	}

	// find distinct piece labels
	keys := make([]string, 0, len(positions))
	for k := range positions {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// validate and create pieces
	pieces := make([]Piece, 0, len(keys))
	for _, k := range keys {
		ps := positions[k]
		if len(ps) < 2 {
			return nil, fmt.Errorf("piece %s length must be >= 2", k)
		}
		stride := ps[1] - ps[0]
		if stride != 1 && stride != w {
			return nil, fmt.Errorf("piece %s has invalid shape", k)
		}
		for i := 1; i < len(ps); i++ {
			if ps[i]-ps[i-1] != stride {
				return nil, fmt.Errorf("piece %s has invalid shape", k)
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
