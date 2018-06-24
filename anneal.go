package rush

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

func anneal(state *Board, maxTemp, minTemp float64, steps int) *Board {
	start := time.Now()
	factor := -math.Log(maxTemp / minTemp)
	state = state.Copy()
	bestState := state.Copy()
	bestEnergy := state.Energy()
	bestTime := start
	previousEnergy := bestEnergy
	rate := steps / 1000
	for step := 0; step < steps; step++ {
		pct := float64(step) / float64(steps-1)
		temp := maxTemp * math.Exp(factor*pct)
		if step%rate == 0 {
			showAnnealProgress(
				step, steps, temp, bestEnergy, time.Since(start).Seconds())
		}
		undo := state.Mutate()
		energy := state.Energy()
		change := energy - previousEnergy
		if change > 0 && math.Exp(-change/temp) < rand.Float64() {
			undo()
		} else {
			previousEnergy = energy
			if energy < bestEnergy {
				bestEnergy = energy
				bestState = state.Copy()
				bestTime = time.Now()
			}
		}
		if time.Since(bestTime).Seconds() > 15 {
			fmt.Println()
			return bestState
		}
	}
	showAnnealProgress(
		steps, steps, minTemp, bestEnergy, time.Since(start).Seconds())
	fmt.Println()
	return bestState
}

func showAnnealProgress(i, n int, t, e, d float64) {
	pct := int(100 * float64(i) / float64(n))
	fmt.Printf("  %3d%% [", pct)
	for p := 0; p < 100; p += 3 {
		if pct > p {
			fmt.Print("=")
		} else {
			fmt.Print(" ")
		}
	}
	fmt.Printf("] %.1f %.2f %.3fs    \r", t, e, d)
}
