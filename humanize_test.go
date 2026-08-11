package rush

import (
	"sort"
	"testing"
)

var humanizeBoards = []string{
	"IBBxooIooLDDJAALooJoKEEMFFKooMGGHHHM",
	"BBoKMxDDDKMoIAALooIoJLEEooJFFNoGGoxN",
	"ooBBMxDDDKMoAAJKoNooJEENIFFLooIGGLox",
	"BBBoKLCCHoKLAAHooMooIDDMGoIJEEGFFJoo",
	"ooIxCCooIDDDooIAAKEEFFoKoooJGGHHHJoo",
	"GHBBxKGHDDDKAAIooLooIEELFFFoJoooooJo",
	"oHBBCCoHoJKoAAoJKLGDDDoLGoIEELFFIooo",
	"BBKooNIJKCCNIJAAMODDoLMOooxLFFGGooxo",
}

func multiset(moves []Move) []string {
	result := make([]string, len(moves))
	for i, move := range moves {
		result[i] = move.String()
	}
	sort.Strings(result)
	return result
}

// TestHumanize confirms that reordering a solution preserves the solution:
// same moves, same number of them, still legal, still solves the board.
func TestHumanize(t *testing.T) {
	for _, desc := range humanizeBoards {
		board, err := NewBoardFromString(desc)
		if err != nil {
			t.Fatalf("%s: %v", desc, err)
		}
		solution := board.Solve()
		if !solution.Solvable {
			t.Fatalf("%s: not solvable", desc)
		}

		moves := board.Humanize(solution.Moves)

		if err := board.ValidateSolution(moves); err != nil {
			t.Errorf("%s: %v", desc, err)
			continue
		}
		before, after := multiset(solution.Moves), multiset(moves)
		if len(before) != len(after) {
			t.Errorf("%s: %d moves became %d", desc, len(before), len(after))
			continue
		}
		for i := range before {
			if before[i] != after[i] {
				t.Errorf("%s: moves differ (%s vs %s)", desc, before[i], after[i])
				break
			}
		}
	}
}

// TestHumanizeIsStable confirms that humanizing an already humanized solution
// changes nothing, so the ordering is a fixed point rather than something that
// keeps drifting each time it is applied.
func TestHumanizeIsStable(t *testing.T) {
	for _, desc := range humanizeBoards {
		board, err := NewBoardFromString(desc)
		if err != nil {
			t.Fatalf("%s: %v", desc, err)
		}
		moves := board.Humanize(board.Solve().Moves)
		again := board.Humanize(moves)
		for i := range moves {
			if moves[i] != again[i] {
				t.Errorf("%s: move %d changed on the second pass", desc, i+1)
				break
			}
		}
	}
}

// TestHumanizeGroupsSubgoals is the point of the whole exercise. A has to get
// past B and C, B can't move until D does, and C can't move until E does. That
// makes two independent two-move subgoals, so D+2 B-1 E+2 C+2 A+4 and
// D+2 E+2 B-1 C+2 A+4 both solve the board - but only the first one finishes a
// thought before starting the next.
func TestHumanizeGroupsSubgoals(t *testing.T) {
	board, err := NewBoard([]string{
		"..DD..",
		"..B...",
		"AABC..",
		"...C..",
		"......",
		"..EE..",
	})
	if err != nil {
		t.Fatal(err)
	}

	interleaved := []Move{{3, 2}, {4, 2}, {1, -1}, {2, 2}, {0, 4}}
	if err := board.ValidateSolution(interleaved); err != nil {
		t.Fatalf("test case is wrong: %v", err)
	}
	expected := []Move{{3, 2}, {1, -1}, {4, 2}, {2, 2}, {0, 4}}

	moves := board.Humanize(interleaved)
	if err := board.ValidateSolution(moves); err != nil {
		t.Fatalf("%v", err)
	}
	for i := range expected {
		if moves[i] != expected[i] {
			t.Fatalf("got %v, want %v", moves, expected)
		}
	}
	if got := board.MoveStretch(moves); got >= board.MoveStretch(interleaved) {
		t.Errorf("stretch did not improve: %d", got)
	}
}
