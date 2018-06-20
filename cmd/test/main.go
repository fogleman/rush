package main

import (
	"fmt"
	"image"
	"log"
	"time"

	"github.com/fogleman/gg"
	"github.com/fogleman/rush"
)

const (
	CellSize = 160
	Padding  = 32
)

const (
	BackgroundColor   = "FFFFFF"
	BoardColor        = "F2EBC7"
	GridLineColor     = "343642"
	PrimaryPieceColor = "962D3E"
	PieceColor        = "348899"
	PieceOutlineColor = "222222"
)

func RenderBoard(board *rush.Board) image.Image {
	const S = CellSize
	bw := board.Width
	bh := board.Height
	w := bw * S
	h := bh * S
	dc := gg.NewContext(w+Padding*2, h+Padding*2)
	dc.Translate(Padding, Padding)
	dc.SetHexColor(BackgroundColor)
	dc.Clear()
	dc.SetHexColor(BoardColor)
	dc.DrawRectangle(0, 0, float64(w+1), float64(h+1))
	dc.Fill()
	for x := S; x < w; x += S {
		fx := float64(x)
		dc.DrawLine(fx, 0, fx, float64(h))
	}
	for y := S; y < h; y += S {
		fy := float64(y)
		dc.DrawLine(0, fy, float64(w), fy)
	}
	dc.SetHexColor(GridLineColor)
	dc.SetLineWidth(2)
	dc.Stroke()
	dc.DrawRectangle(0, 0, float64(w+1), float64(h+1))
	dc.SetLineWidth(6)
	dc.Stroke()
	for i, piece := range board.Pieces {
		stride := 1
		if piece.Direction == rush.Vertical {
			stride = bw
		}
		i0 := piece.Position
		i1 := i0 + stride*(piece.Size-1)
		x0 := float64(i0 % bw)
		y0 := float64(i0 / bw)
		x1 := float64(i1 % bw)
		y1 := float64(i1 / bw)
		dx := x1 - x0
		dy := y1 - y0
		m := S / 8.0
		px := x0*S + m
		py := y0*S + m
		pw := dx*S + S - m*2
		ph := dy*S + S - m*2
		dc.DrawRoundedRectangle(px+0.5, py+0.5, pw, ph, S/8.0)
		if i == 0 {
			dc.SetHexColor(PrimaryPieceColor)
		} else {
			dc.SetHexColor(PieceColor)
		}
		dc.FillPreserve()
		dc.SetLineWidth(S / 32.0)
		dc.SetHexColor(PieceOutlineColor)
		dc.Stroke()
	}
	return dc.Image()
}

func main() {
	desc := []string{
		"BBBCDE",
		"FGGCDE",
		"F.AADE",
		"HHI...",
		".JI.KK",
		".JLLMM",
	}
	board, err := rush.NewBoard(desc)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(board)
	fmt.Println()

	// gg.SavePNG(fmt.Sprintf("%04d.png", 0), RenderBoard(board))

	start := time.Now()
	moves, ok := board.Solve(16)
	if !ok {
		log.Fatal("no solution")
	}
	elapsed := time.Since(start)

	fmt.Println(moves)
	sum := 0
	for _, move := range moves {
		board.DoMove(move)
		fmt.Println(move)
		fmt.Println(board)
		fmt.Println()
		sum += move.AbsSteps()
		// gg.SavePNG(fmt.Sprintf("%04d.png", i+1), RenderBoard(board))
	}
	fmt.Println(len(moves))
	fmt.Println(sum)
	fmt.Println(elapsed)
}
