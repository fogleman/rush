package rush

import (
	"image"

	"github.com/fogleman/gg"
)

const (
	CellSize = 160
	Padding  = 32
)

const (
	BackgroundColor   = "FFFFFF"
	BoardColor        = "F2EBC7"
	BlockedColor      = "D96D60"
	GridLineColor     = "343642"
	PrimaryPieceColor = "962D3E"
	PieceColor        = "348899"
	PieceOutlineColor = "222222"
	LabelColor        = "222222"
	WallColor         = "111111"
)

const Font = "/Library/Fonts/Arial.ttf"

func renderBoard(board *Board) image.Image {
	const S = CellSize
	bw := board.Width
	bh := board.Height
	w := bw * S
	h := bh * S
	dc := gg.NewContext(w+Padding*2, h+Padding*2)
	dc.LoadFontFace(Font, 36)
	dc.Translate(Padding, Padding)
	dc.SetHexColor(BackgroundColor)
	dc.Clear()
	dc.SetHexColor(BoardColor)
	dc.DrawRectangle(0, 0, float64(w+1), float64(h+1))
	dc.Fill()
	for _, i := range board.BlockedSquares() {
		x := float64(i % bw)
		y := float64(i / bw)
		dc.DrawRectangle(x*S, y*S, S, S)
	}
	dc.SetHexColor(BlockedColor)
	dc.Fill()
	p := S / 8.0
	r := S / 32.0
	for _, i := range board.Walls {
		x := float64(i % bw)
		y := float64(i / bw)
		dc.DrawRectangle(x*S, y*S, S, S)
		dc.SetHexColor(WallColor)
		dc.Fill()
		dc.DrawCircle(x*S+p, y*S+p, r)
		dc.DrawCircle(x*S+S-p, y*S+p, r)
		dc.DrawCircle(x*S+p, y*S+S-p, r)
		dc.DrawCircle(x*S+S-p, y*S+S-p, r)
		dc.SetHexColor(BoardColor)
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
	dc.SetHexColor(GridLineColor)
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
			dc.SetHexColor(PrimaryPieceColor)
		} else {
			dc.SetHexColor(PieceColor)
		}
		dc.FillPreserve()
		dc.SetLineWidth(S / 32.0)
		dc.SetHexColor(PieceOutlineColor)
		dc.Stroke()
		tx := px + pw/2
		ty := py + ph/2
		dc.SetHexColor(LabelColor)
		dc.DrawStringAnchored(string('A'+i), tx, ty, 0.5, 0.5)
	}
	return dc.Image()
}
