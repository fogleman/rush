package rush

import (
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/fogleman/gg"
)

const (
	showBlockedSquares = true
	showPieceLabels    = true
	showSolution       = false
)

const (
	cellSize = 160
	padding  = 32
)

const (
	backgroundColor   = "FFFFFF"
	boardColor        = "F2EBC7"
	blockedColor      = "D96D60"
	gridLineColor     = "343642"
	primaryPieceColor = "962D3E"
	pieceColor        = "348899"
	pieceOutlineColor = "222222"
	labelColor        = "222222"
	wallColor         = "111111"
)

const labelFont = "/Library/Fonts/Arial.ttf"
const footerFont = "/System/Library/Fonts/Menlo.ttc"

func renderBoard(board *Board) image.Image {
	const S = cellSize
	bw := board.Width
	bh := board.Height
	w := bw * S
	h := bh * S
	iw := w + padding*2
	ih := h + padding*2
	if showSolution {
		ih += 88
	}
	dc := gg.NewContext(iw, ih)
	dc.LoadFontFace(labelFont, 36)
	dc.Translate(padding, padding)
	dc.SetHexColor(backgroundColor)
	dc.Clear()
	dc.SetHexColor(boardColor)
	dc.DrawRectangle(0, 0, float64(w+1), float64(h+1))
	dc.Fill()
	if showBlockedSquares {
		for _, i := range board.BlockedSquares() {
			x := float64(i % bw)
			y := float64(i / bw)
			dc.DrawRectangle(x*S, y*S, S, S)
		}
		dc.SetHexColor(blockedColor)
		dc.Fill()
	}
	p := S / 8.0
	r := S / 32.0
	for _, i := range board.Walls {
		x := float64(i % bw)
		y := float64(i / bw)
		dc.DrawRectangle(x*S, y*S, S, S)
		dc.SetHexColor(wallColor)
		dc.Fill()
		dc.DrawCircle(x*S+p, y*S+p, r)
		dc.DrawCircle(x*S+S-p, y*S+p, r)
		dc.DrawCircle(x*S+p, y*S+S-p, r)
		dc.DrawCircle(x*S+S-p, y*S+S-p, r)
		dc.SetHexColor(boardColor)
		dc.Fill()
	}
	for x := S; x < w; x += S {
		fx := float64(x)
		dc.DrawLine(fx, 0, fx, float64(h))
	}
	for y := S; y < h; y += S {
		fy := float64(y)
		dc.DrawLine(0, fy, float64(w), fy)
	}
	dc.SetHexColor(gridLineColor)
	dc.SetLineWidth(2)
	dc.Stroke()
	dc.DrawRectangle(0, 0, float64(w+1), float64(h+1))
	dc.SetLineWidth(6)
	dc.Stroke()
	for i, piece := range board.Pieces {
		stride := 1
		if piece.Orientation == Vertical {
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
			dc.SetHexColor(primaryPieceColor)
		} else {
			dc.SetHexColor(pieceColor)
		}
		dc.FillPreserve()
		dc.SetLineWidth(S / 32.0)
		dc.SetHexColor(pieceOutlineColor)
		dc.Stroke()
		if showPieceLabels {
			tx := px + pw/2
			ty := py + ph/2
			dc.SetHexColor(labelColor)
			dc.DrawStringAnchored(string('A'+i), tx, ty, 0.5, 0.5)
		}
	}

	if showSolution {
		x := float64(w) / 2
		y := float64(h) + padding*0.75
		footer := ""
		solution := board.Solve()
		if solution.Solvable {
			moveStrings := make([]string, len(solution.Moves))
			for i, move := range solution.Moves {
				moveStrings[i] = move.String()
			}
			footer = fmt.Sprintf("%s (%d moves)",
				strings.Join(moveStrings, " "), solution.NumMoves)
		}
		dc.LoadFontFace(footerFont, 24)
		var tw float64
		for _, line := range dc.WordWrap(footer, float64(w)) {
			w, _ := dc.MeasureString(line)
			tw = math.Max(tw, w)
		}
		dc.DrawStringWrapped(footer, x, y, 0.5, 0, tw, 1.5, gg.AlignLeft)
	}

	return dc.Image()
}
