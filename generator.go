package rush

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

type Generator struct {
	Width    int
	Height   int
	Target   int
	Pieces   []Piece
	Occupied []bool
}

func NewGenerator(w, h, target, size int, orientation Orientation) *Generator {
	var pieces []Piece
	if orientation == Horizontal {
		y := target / w
		pieces = append(pieces, Piece{y * w, size, orientation})
	} else {
		x := target % w
		pieces = append(pieces, Piece{x, size, orientation})
	}
	occupied := make([]bool, w*h)
	for _, piece := range pieces {
		updateOccupied(occupied, piece.Stride(w), piece.Position, piece.Size, true)
	}
	return &Generator{w, h, target, pieces, occupied}
}

func (g *Generator) Copy() *Generator {
	return g
}

func (g *Generator) Energy() float64 {
	board := Board{g.Width, g.Height, g.Pieces, g.Occupied}
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
