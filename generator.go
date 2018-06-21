package rush

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type GeneratorConfig struct {
	Width              int
	Height             int
	Target             int
	PrimarySize        int
	PrimaryOrientation Orientation
	MinPieces          int
	MaxPieces          int
	MinSize            int
	MaxSize            int
}

// TODO: piece pool / bag - in use / out of use pieces

type Generator struct {
	Width    int
	Height   int
	Target   int
	Pieces   []Piece
	Occupied []bool
}

func NewGenerator(w, h, target, size int, orientation Orientation) *Generator {
	var pieces []Piece
	// place the primary piece
	if orientation == Horizontal {
		y := target / w
		pieces = append(pieces, Piece{y * w, size, orientation})
	} else {
		x := target % w
		pieces = append(pieces, Piece{x, size, orientation})
	}
	// create occupied grid
	occupied := make([]bool, w*h)
	for _, piece := range pieces {
		updateOccupied(occupied, piece.Stride(w), piece.Position, piece.Size, true)
	}
	return &Generator{w, h, target, pieces, occupied}
}

func (g *Generator) Generate() *Board {
	n := rand.Intn(12) + 3
	for i := 0; i < n; i++ {
		piece, ok := g.randomPiece(100)
		if ok {
			g.Pieces = append(g.Pieces, piece)
			updateOccupied(g.Occupied, piece.Stride(g.Width), piece.Position, piece.Size, true)
		}
	}
	return &Board{g.Width, g.Height, g.Pieces, g.Occupied, MakeMemoKey(g.Pieces)}
}

func (g *Generator) Copy() *Generator {
	pieces := make([]Piece, len(g.Pieces))
	occupied := make([]bool, len(g.Occupied))
	copy(pieces, g.Pieces)
	copy(occupied, g.Occupied)
	return &Generator{g.Width, g.Height, g.Target, pieces, occupied}
}

func (g *Generator) Energy() float64 {
	board := Board{g.Width, g.Height, g.Pieces, g.Occupied, MakeMemoKey(g.Pieces)}
	moves, ok := board.Solve(g.Target)
	if !ok {
		return 1
	}
	return -float64(len(moves))
}

func (g *Generator) DoMove() {
	// do a random move
	// add a piece
	// remove a piece
	// remove & add a piece
}

func (g *Generator) UndoMove() {
}

func (g *Generator) randomPiece(maxAttempts int) (Piece, bool) {
	w := g.Width
	h := g.Height
	for i := 0; i < maxAttempts; i++ {
		size := 2 + rand.Intn(2) // TODO: weighted
		orientation := Orientation(rand.Intn(2))
		var x, y, stride int
		if orientation == Vertical {
			x = rand.Intn(w)
			y = rand.Intn(h - size + 1)
			stride = w
		} else {
			x = rand.Intn(w - size + 1)
			y = rand.Intn(h)
			stride = 1
		}
		position := y*w + x
		idx := position
		ok := true
		for j := 0; j < size; j++ {
			if g.Occupied[idx] {
				ok = false
				break
			}
			idx += stride
		}
		if ok {
			return Piece{position, size, orientation}, true
		}
	}
	return Piece{}, false
}

func anneal(state *Generator, maxTemp, minTemp float64, steps int) *Generator {
	start := time.Now()
	rate := steps / 200
	factor := -math.Log(maxTemp / minTemp)
	state = state.Copy()
	bestState := state.Copy()
	bestEnergy := state.Energy()
	previousEnergy := bestEnergy
	for step := 0; step < steps; step++ {
		pct := float64(step) / float64(steps-1)
		temp := maxTemp * math.Exp(factor*pct)
		if step%rate == 0 {
			showProgress(step, steps, bestEnergy, time.Since(start).Seconds())
		}
		state.DoMove()
		energy := state.Energy()
		change := energy - previousEnergy
		if change > 0 && math.Exp(-change/temp) < rand.Float64() {
			state.UndoMove()
		} else {
			previousEnergy = energy
			if energy < bestEnergy {
				bestEnergy = energy
				bestState = state.Copy()
			}
		}
	}
	showProgress(steps, steps, bestEnergy, time.Since(start).Seconds())
	fmt.Println()
	return bestState
}

func showProgress(i, n int, e, d float64) {
	pct := int(100 * float64(i) / float64(n))
	fmt.Printf("  %3d%% [", pct)
	for p := 0; p < 100; p += 3 {
		if pct > p {
			fmt.Print("=")
		} else {
			fmt.Print(" ")
		}
	}
	fmt.Printf("] %.6f %.3fs    \r", e, d)
}
