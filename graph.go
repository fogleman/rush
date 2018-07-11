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
	linksToNonSolved := make(map[int]bool)

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
	for _, move := range solution.Moves {
		solutionIDs = append(solutionIDs, ids[*board.MemoKey()])
		board.DoMove(move)
	}
	solutionIDs = append(solutionIDs, ids[*board.MemoKey()])

	links := make(map[Link]bool)
	solutionLinks := make(map[Link]bool)
	for j := 1; j < len(solutionIDs); j++ {
		i := j - 1
		link := MakeLink(solutionIDs[i], solutionIDs[j])
		solutionLinks[link] = true
	}

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
				numMoves1 := idToNumMoves[id1]
				numMoves2 := idToNumMoves[id2]
				if numMoves1 != 0 || numMoves2 != 0 {
					linksToNonSolved[id1] = true
					linksToNonSolved[id2] = true
				}
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

	for id := 0; id < idCounter; id++ {
		if !linksToNonSolved[id] {
			continue
		}
		fmt.Printf("%d [label=\"\" shape=circle style=filled fillcolor=\"#EDD569\"];\n", id)
	}

	for _, link := range sortedLinks {
		a := link.Src
		b := link.Dst
		weight := 1
		if _, ok := solutionLinks[link]; ok {
			weight = 100
		}
		if !linksToNonSolved[a] || !linksToNonSolved[b] {
			continue
		}
		if idToNumMoves[a] > idToNumMoves[b] {
			fmt.Printf("%d -> %d [arrowsize=0.5, weight=%d];\n", a, b, weight)
		} else if idToNumMoves[a] < idToNumMoves[b] {
			fmt.Printf("%d -> %d [arrowsize=0.5, weight=%d];\n", b, a, weight)
		} else {
			fmt.Printf("%d -> %d [constraint=false, arrowhead=none, color=\"#00000020\"];\n", a, b)
		}
	}
	for _, ids := range numMovesToIDs {
		fmt.Printf("{ rank=same; ")
		first := true
		for _, id := range ids {
			if !linksToNonSolved[id] {
				continue
			}
			if !first {
				fmt.Printf(", ")
			}
			fmt.Printf("%d", id)
			first = false
		}
		fmt.Println(" }")
	}
	for _, id := range solutionIDs {
		fmt.Printf("%d [style=filled, fillcolor = \"#3F628F\"];\n", id)
	}
	for _, id := range numMovesToIDs[maxMoves] {
		fmt.Printf("%d [style=filled, fillcolor = \"#E94128\"];\n", id)
	}
	for _, id := range numMovesToIDs[0] {
		if !linksToNonSolved[id] {
			continue
		}
		fmt.Printf("%d [style=filled, fillcolor = \"#458955\"];\n", id)
	}
	fmt.Println("}")
}
