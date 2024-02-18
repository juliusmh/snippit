package elements

import (
	"fmt"
	"image/color"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

// Context for drawing graphics to.
type Context interface {
	DrawRectangle(x, y, w, h, r float64, c color.Color)
	StrokeRectangle(x, y, w, h, r, sw float64, c color.Color)
	DrawCircle(x, y, r float64, c color.Color)
	DrawText(x, y float64, s string, c color.Color, f font.Face)
}

// Element that can be rendered to the screen.
type Element interface {
	Draw(ctx Context, x, y float64)
	Measure() (w, h float64)
}

func parseColor(s string) (color.Color, error) {
	var c color.RGBA
	var err error
	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4")

	}
	return c, err
}

func mustParseFont(data []byte) *truetype.Font {
	ttf, err := truetype.Parse(fontBytes)
	if err != nil {
		panic(fmt.Errorf("failed to parse font: %+v", err))
	}
	return ttf
}
