package elements

import (
	_ "embed"
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

//go:embed fonts/roboto_mono_regular.ttf
var fontBytes []byte

// RobotoMono is the default font face.
var RobotoMono *truetype.Font

func init() {
	RobotoMono = mustParseFont(fontBytes)
}

type Code struct {
	Text            string
	Face            font.Face
	ShowLineNumbers bool
	LineNumberColor color.Color
	Lexer           chroma.Lexer
	Style           *chroma.Style
}

func (c Code) Measure() (float64, float64) {
	lines, columns := 0.0, 0.0
	i := 0.0
	for _, r := range c.Text {
		if r == '\n' {
			columns = max(columns, i)
			lines++
			i = 0
		} else {
			i += c.charWidth(r)
		}
	}

	lo, ro := c.lineNumberOffset()
	columns += lo + ro

	ch := c.lineHeight()
	return columns, lines * ch
}

func (c Code) Draw(ctx Context, x, y float64) {
	iterator, err := c.Lexer.Tokenise(nil, c.Text)
	if err != nil {
		panic(fmt.Errorf("could not tokenize code: %+v", err))
	}

	// Draw the code:
	lineHeight := c.lineHeight()
	lnLeftOff, lnRightOff := c.lineNumberOffset()
	atX, atY := x+lnLeftOff+lnRightOff, y
	for t := iterator(); t != chroma.EOF; t = iterator() {
		col := c.Style.Get(t.Type).Colour
		for _, r := range t.Value {
			if r == '\n' {
				atX = x + lnLeftOff + lnRightOff
				atY += lineHeight
			} else {
				ctx.DrawText(atX, atY, string(r), color.RGBA{
					R: col.Red(),
					G: col.Green(),
					B: col.Blue(),
					A: 0xFF,
				}, c.Face)
				atX += c.charWidth(r)
			}
		}
	}

	// Draw the line numbers:
	if c.ShowLineNumbers {

		// TODO: this needs cleanup:
		lineCount := strings.Count(c.Text, "\n")
		cw := c.charWidth('9')

		for i := 0; i < lineCount; i++ {
			digits := math.Ceil(math.Log10(2 + float64(i)))
			atX = x + lnLeftOff/2 - (digits*cw)/2
			atY = y + float64(i)*lineHeight
			l := strconv.Itoa(i + 1)
			ctx.DrawText(atX, atY, l, c.LineNumberColor, c.Face)
		}

	}

}

func (c Code) lineHeight() float64 {
	metrics := c.Face.Metrics()
	h := metrics.Height.Round()
	return float64(h) + 2
}

func (c Code) charWidth(r rune) float64 {
	advance, ok := c.Face.GlyphAdvance(r)
	if !ok {
		panic(fmt.Errorf("unknown rune: %s", string(r)))
	}
	return float64(advance.Round())
}

func (c Code) lineNumberOffset() (float64, float64) {
	if !c.ShowLineNumbers {
		return 0, 0
	}
	cw := c.charWidth('9')
	lines := strings.Count(c.Text, "\n")
	digits := math.Ceil(math.Log10(float64(lines)))
	return (digits + 2) * cw, 1 * cw
}
