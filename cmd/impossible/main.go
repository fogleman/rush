package main

import (
	"fmt"
	"sort"

	"github.com/fogleman/gg"
	"github.com/fogleman/rush"
)

type Key [rush.MaxPieces]rush.Piece

func makeKey(board *rush.Board) Key {
	pieces := make([]rush.Piece, len(board.Pieces))
	copy(pieces, board.Pieces)
	sort.Slice(pieces, func(i, j int) bool {
		if i == 0 {
			return true
		}
		a := pieces[i]
		b := pieces[j]
		if a.Orientation != b.Orientation {
			return a.Orientation < b.Orientation
		}
		if a.Size != b.Size {
			return a.Size < b.Size
		}
		return a.Position < b.Position
	})
	var key Key
	for i, piece := range pieces {
		key[i] = piece
	}
	return key
}

func main() {
	seen := make(map[Key]bool)
	counter := 0
	for i := 0; ; i++ {
		board := rush.NewRandomBoard(6, 6, 2, 2, 4, 0)
		if board.Impossible() {
			continue
		}
		key := makeKey(board)
		if _, ok := seen[key]; ok {
			continue
		}
		// fmt.Println(key)
		seen[key] = true
		if board.Validate() != nil {
			continue
		}
		solution := board.Solve()
		if solution.Solvable {
			continue
		}
		gg.SavePNG(fmt.Sprintf("impossible-%d.png", counter), board.Render())
		counter++
		fmt.Println(counter)
	}
}
