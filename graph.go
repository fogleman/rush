package rush

import (
	"fmt"
	"sort"
)

type Link struct {
	Src, Dst int
}

func MakeLink(id1, id2 int) Link {
	if id1 > id2 {
		id1, id2 = id2, id1
	}
	return Link{id1, id2}
}

func Graph(input *Board) {
	ids := make(map[MemoKey]int)
	numMovesToIDs := make(map[int][]int)
	idToNumMoves := make(map[int]int)
	var solutionIDs []int
	idCounter := 0

	var q []*Board
	q = append(q, input)

	fmt.Println("digraph g {")
	fmt.Println("pad=1;")
	var maxMoves int
	solver := NewSolverWithStaticAnalyzer(input, nil)
	for len(q) > 0 {
		board := q[len(q)-1]
		key := board.MemoKey()
		q = q[:len(q)-1]
		if _, ok := ids[*key]; ok {
			continue
		}
		id := idCounter
		idCounter++
		ids[*key] = id
		solver.board = board
		numMoves := solver.UnsafeSolve().NumMoves
		if numMoves > maxMoves {
			maxMoves = numMoves
		}
		numMovesToIDs[numMoves] = append(numMovesToIDs[numMoves], id)
		idToNumMoves[id] = numMoves
		moves := board.Moves(nil)
		for _, move := range moves {
			board.DoMove(move)
			q = append(q, board.Copy())
			board.UndoMove(move)
		}
	}

	board, solution := input.Unsolve()
	board.DoMove(solution.Moves[0])
	for _, move := range solution.Moves[1:] {
		solutionIDs = append(solutionIDs, ids[*board.MemoKey()])
		board.DoMove(move)
	}

	for id := 0; id < idCounter; id++ {
		fmt.Printf("%d [label=\"\" shape=circle style=filled fillcolor=\"#EDD569\"];\n", id)
	}

	links := make(map[Link]bool)

	q = append(q, input)
	for len(q) > 0 {
		board := q[len(q)-1]
		key := board.MemoKey()
		q = q[:len(q)-1]
		id1 := ids[*key]
		moves := board.Moves(nil)
		for _, move := range moves {
			board.DoMove(move)
			id2 := ids[*board.MemoKey()]
			link := MakeLink(id1, id2)
			if _, ok := links[link]; !ok {
				links[link] = true
				q = append(q, board.Copy())
			}
			board.UndoMove(move)
		}
	}

	var sortedLinks []Link
	for link := range links {
		sortedLinks = append(sortedLinks, link)
	}

	sort.Slice(sortedLinks, func(i, j int) bool {
		a, b := sortedLinks[i], sortedLinks[j]
		if a.Src == b.Src {
			return a.Dst < b.Dst
		}
		return a.Src < b.Src
	})

	for _, link := range sortedLinks {
		a := link.Src
		b := link.Dst
		if idToNumMoves[a] > idToNumMoves[b] {
			fmt.Printf("%d -> %d [arrowsize=0.5];\n", a, b)
		} else if idToNumMoves[a] < idToNumMoves[b] {
			fmt.Printf("%d -> %d [arrowsize=0.5];\n", b, a)
		} else {
			fmt.Printf("%d -> %d [constraint=false, arrowhead=none, color=\"#00000020\"];\n", a, b)
		}
	}
	for _, ids := range numMovesToIDs {
		fmt.Printf("{ rank=same; ")
		for i, id := range ids {
			if i != 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%d", id)
		}
		fmt.Println(" }")
	}
	for _, id := range numMovesToIDs[maxMoves] {
		fmt.Printf("%d [style=filled, fillcolor = \"#E94128\"];\n", id)
	}
	for _, id := range numMovesToIDs[0] {
		fmt.Printf("%d [style=filled, fillcolor = \"#458955\"];\n", id)
	}
	for _, id := range solutionIDs {
		fmt.Printf("%d [style=filled, fillcolor = \"#3F628F\"];\n", id)
	}
	fmt.Println("}")
}
