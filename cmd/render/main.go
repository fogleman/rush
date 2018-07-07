package main

import (
	"fmt"
	"log"
	"os"

	"github.com/fogleman/gg"
	"github.com/fogleman/rush"
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 && len(args) != 2 {
		fmt.Println("render DESC [OUTPUT]")
		return
	}

	board, err := rush.NewBoardFromString(args[0])
	if err != nil {
		log.Fatal(err)
	}

	output := "out.png"
	if len(args) == 2 {
		output = args[1]
	}
	err = gg.SavePNG(output, board.Render())
	if err != nil {
		log.Fatal(err)
	}
}
