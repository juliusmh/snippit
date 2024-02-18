package elements

import "image/color"

type Dots struct {
	// Radius of the individual dots.
	Radius float64
	// Spacing between dots. Zero meaning dots touch each other.
	Spacing float64
}

func (d Dots) Draw(ctx Context, x, y float64) {
	colors := []color.Color{
		color.RGBA{R: 0xFF, G: 0x5F, B: 0x56, A: 255},
		color.RGBA{R: 0xFF, G: 0xBD, B: 0x2E, A: 255},
		color.RGBA{R: 0x27, G: 0xC9, B: 0x3F, A: 255},
	}
	atX := x + d.Radius
	atY := y + d.Radius
	for _, col := range colors {
		ctx.DrawCircle(atX, atY, d.Radius, col)
		atX += 2*d.Radius + d.Spacing
	}
}

func (d Dots) Measure() (float64, float64) {
	return 3 * (2*d.Radius + d.Spacing), 2 * d.Radius
}

type FileName struct {
	Text  string
	Color color.Color
}

func (f FileName) Draw(ctx Context, x, y float64) {
	//TODO implement me
	panic("implement me")
}

func (f FileName) Measure() (w, h float64) {
	//TODO implement me
	panic("implement me")
}
