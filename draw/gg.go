package draw

import (
	"image/color"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

type GGContext struct {
	*gg.Context
}

func (g GGContext) StrokeRectangle(x, y, w, h, r, sw float64, c color.Color) {
	g.Context.SetColor(c)
	g.Context.DrawRoundedRectangle(
		x, y, w, h, r,
	)
	g.Context.SetLineWidth(sw)
	g.Context.Stroke()
}

func GG(w, h, scale float64) GGContext {
	dc := gg.NewContext(int(w*scale), int(h*scale))
	return GGContext{Context: dc}
}

func (g GGContext) DrawRectangle(x, y, w, h, r float64, c color.Color) {
	g.Context.SetColor(c)
	g.Context.DrawRoundedRectangle(x, y, w, h, r)
	g.Context.Fill()
}

func (g GGContext) DrawCircle(x, y, r float64, c color.Color) {
	g.Context.SetColor(c)
	g.Context.DrawCircle(x, y, r)
	g.Context.Fill()
}

func (g GGContext) DrawText(x, y float64, r string, c color.Color, f font.Face) {
	g.Context.SetFontFace(f)
	g.Context.SetColor(c)
	g.Context.DrawStringAnchored(r, x, y, 0, 1)
}
