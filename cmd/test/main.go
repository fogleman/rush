package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/fogleman/rush"
)

func main() {
	desc := []string{
		"....CE",
		"..BBCE",
		"..AADE",
		"..H.D.",
		"..HGFF",
		"..HG..",
	}
	board, err := rush.NewBoard(desc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(board)
	fmt.Println()

	for i := 0; i < 10; i++ {
		moves := board.Moves()
		fmt.Println(len(moves), moves)
		move := moves[rand.Intn(len(moves))]
		board.DoMove(move)
		fmt.Println(move)
		fmt.Println()
		fmt.Println(board)
		fmt.Println()
	}
}
