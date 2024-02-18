package elements

import "image/color"

type Card struct {
	BackgroundColor color.Color
	BorderColor     color.Color
	BorderRadius    float64
	BorderWidth     float64
	Body            Element
}

func (c Card) Measure() (float64, float64) {
	return c.Body.Measure()
}

func (c Card) Draw(ctx Context, x, y float64) {
	w, h := c.Body.Measure()
	ctx.DrawRectangle(
		x,
		y,
		w,
		h,
		c.BorderRadius,
		c.BackgroundColor,
	)
	ctx.StrokeRectangle(
		x,
		y,
		w,
		h,
		c.BorderRadius,
		c.BorderWidth,
		c.BorderColor,
	)
	c.Body.Draw(
		ctx,
		x,
		y,
	)
}
