package main

import (
	"fmt"
	"log"

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
}
