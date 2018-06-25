# Rush Hour Solver

This is a puzzle solver and generator for [Rush Hour](https://en.wikipedia.org/wiki/Rush_Hour_(puzzle)).
You may also know this game from one of its iOS implementations, such as [Unblock Me](https://itunes.apple.com/us/app/unblock-me/id315019111?mt=8).

The code is written in [Go](https://golang.org/).

![Examples](https://i.imgur.com/YlT8Y39.png)

### Features

- Can solve puzzles
- Can generate puzzles (such as those shown above)
- Supports arbitrarily-sized boards
- Supports arbitrarily-sized pieces
- Supports "walls" (immovable obstacles)
- Uses "static analysis" to quickly determine if a puzzle cannot be solved without performing a full search
- Can "unsolve" puzzles - given a starting state, find a reachable state that is furthest from the win state
- Can render puzzles to PNG

### Example Puzzle Solution

The solver found a solution to this puzzle in 75 milliseconds. It requires 50 moves (82 steps). Only 3,519 distinct states were found while searching this puzzle. But the memoization cache was queried 595,093 times.

![Example](https://i.imgur.com/eWnPtLo.gif)

### ASCII Format

Puzzles are displayed and can be parsed in a simple ASCII format:

```
BBBCDE
FGGCDE
F.AADE
HHI...
.JI.KK
.JLLMM
```

Empty cells are indicated with a `.`. Walls are indicated with a lowercase `x`. Pieces are indicated with capital letters starting with `A`.

The "primary" piece (the red car) is labeled `A` and must always be horizontal, but can appear on any row. It will always exit to the right. No other horizontal pieces may be on that same row.

### API Example

```go
// define the puzzle in ASCII
desc := []string{
	"BBBCDE",
	"FGGCDE",
	"F.AADE",
	"HHI...",
	".JI.KK",
	".JLLMM",
}

// parse and create a board
board, err := rush.NewBoard(desc)
if err != nil {
	log.Fatal(err)
}

// compute a solution
solution := board.Solve()

// print out solution information
fmt.Printf("solvable: %t\n", solution.Solvable)
fmt.Printf(" # moves: %d\n", solution.NumMoves)
fmt.Printf(" # steps: %d\n", solution.NumSteps)

// print out moves to solve puzzle
moveStrings := make([]string, len(solution.Moves))
for i, move := range solution.Moves {
	moveStrings[i] = move.String()
}
fmt.Println(strings.Join(moveStrings, ", "))

// solvable: true
//  # moves: 49
//  # steps: 93
// A-1, C+2, B+1, E+1, F-1, A-1, I-1, K-2, D+2, B+2, G+2, I-2, A+1, H+1,
// F+4, A-1, H-1, I+2, B-2, E-1, G-3, C-1, D-2, I-1, H+4, F-1, J-1, K+2,
// L-2, C+3, I+3, A+2, G+2, F-3, H-2, D+1, B+1, J-3, A-2, H-2, C-2, I-2,
// K-4, C+1, I+1, M-2, D+2, E+3, A+4
```

Run it yourself:

```
go get -u github.com/fogleman/rush
cd ~/go/src/github.com/fogleman/rush
go run cmd/example/main.go
```

### Solving

The `Solver` works via an [iterative deepening depth-first search](https://en.wikipedia.org/wiki/Iterative_deepening_depth-first_search) with [memoization](https://en.wikipedia.org/wiki/Memoization) to avoid searching the same position multiple times. Before searching, the `StaticAnalyzer` is invoked to ensure that no cells between the primary piece and its exit are permanently blocked.

### Unsolving

The `Unsolver` takes any existing solvable configuration and tries to make it "harder" (require more moves) by finding some other reachable state that is further from the win.

In the example below, the input puzzle (left) is already solved. The unsolver moves the pieces around (with valid moves only) and produces a puzzle that requires 45 moves to solve (right). Note that these puzzles have the same pieces.

![Unsolving Example](https://i.imgur.com/QNSKKU5.png)

<p align="center"><i>The Unsolver worked backwards to transform an already-solved puzzle into one that requires 45 moves to solve</i></p>

### Generating

The `Generator` creates puzzles via [simulated annealing](https://en.wikipedia.org/wiki/Simulated_annealing). The possible mutations are:

- make a random valid move
- add a piece
- remove a piece
- remove & add (move) a piece
- add a wall
- remove a wall
- remove & add (move) a wall

The score is based on how many moves and how many steps are required to solve the puzzle. Other scoring functions could conceivably generate "interesting" puzzles based on some other metric.

After generating a puzzle, the `Unsolver` is invoked to see if the puzzle can be made a bit harder.

As the puzzle needs to be solved after every iteration during annealing, generating puzzles isn't very fast. I'm still looking into other ways of generating interesting puzzles.

### Static Analysis

The red cells shown in the example below will always be occupied, no matter what sequence of moves are made. This is determined via static analysis. This algorithm can quickly weed out impossible puzzles before performing a more expensive search algorithm. (The search algorithm must explore all reachable states before concluding that the puzzle cannot be solved.) The static analysis doesn't catch every impossible puzzle, of course.

The static analysis algorithm is well-documented. See [static.go](https://github.com/fogleman/rush/blob/master/static.go) to learn more!

![Static Analysis Example](https://i.imgur.com/ZHs3XHp.png)

<p align="center"><i>The static analysis algorithm quickly determined that this puzzle cannot be solved</i></p>

### Sample Puzzles

Below are several sample puzzles created by the simulated annealing algorithm. Their solutions are also provided. Some of these had constraints on the number or size of pieces that could be used.

![Sample Puzzles](https://i.imgur.com/YuEUSmr.png)

### 7x7 Puzzle

Here is an example 7x7 puzzle, just to demonstrate that the code supports arbitrarily sized boards. The red squares indicate cells that will always be occupied no matter what moves are made, as determined by the static analysis.

![7x7](https://i.imgur.com/uyUyyEW.png)
