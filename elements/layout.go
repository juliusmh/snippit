package elements

type Column []Element

func (c Column) Draw(ctx Context, x, y float64) {
	at := y
	for _, e := range c {
		e.Draw(ctx, x, at)
		_, h := e.Measure()
		at += h
	}
}

func (c Column) Measure() (float64, float64) {
	maxW, sumH := 0.0, 0.0
	for _, e := range c {
		w, h := e.Measure()
		maxW = max(w, maxW)
		sumH += h
	}
	return maxW, sumH
}

// Row draws elements in a row.
type Row []Element

func (c Row) Draw(ctx Context, x, y float64) {
	at := x
	for _, e := range c {
		e.Draw(ctx, at, y) // TODO: align vertically: center
		w, _ := e.Measure()
		at += w
	}
}

func (c Row) Measure() (float64, float64) {
	sumW, maxH := 0.0, 0.0
	for _, e := range c {
		w, h := e.Measure()
		sumW += w
		maxH = max(h, maxH)
	}
	return sumW, maxH
}

// Padding adds padding around a component.
type Padding struct {
	Element

	Left, Right float64
	Top, Bottom float64
}

func (c Padding) Draw(ctx Context, x, y float64) {
	startX := x + c.Left
	startY := y + c.Top
	c.Element.Draw(ctx, startX, startY)
}

func (c Padding) Measure() (float64, float64) {
	bodyW, bodyH := c.Element.Measure()
	return bodyW + c.Left + c.Right, bodyH + c.Top + c.Bottom
}
