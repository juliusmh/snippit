package main

import (
	"bytes"
	"embed"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"math"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/esimov/stackblur-go"
	"github.com/golang/freetype/truetype"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d/draw2dkit"
	"github.com/llgcode/draw2d/draw2dsvg"
	"github.com/zyedidia/highlight"
)

var (

	//go:embed syntax/*.yaml
	syntaxFS embed.FS
	//go:embed fonts/*.ttf
	fontFS embed.FS
	//go:embed themes/*.json
	themesFS embed.FS

	red    = color.RGBA{R: 0xFF, G: 0x5F, B: 0x56, A: 255}
	yellow = color.RGBA{R: 0xFF, G: 0xBD, B: 0x2E, A: 255}
	green  = color.RGBA{R: 0x27, G: 0xC9, B: 0x3F, A: 255}
)

func main() {
	var (
		out        = flag.String("out", "snippet.png", "output file to use")
		themeName  = flag.String("theme", "solarized-dark", "theme to use")
		syntaxName = flag.String("syntax", "", "syntax mode to use")

		windowDecorations = flag.Bool("decorations", false, "use window decorations")
		dropShadow        = flag.Bool("shadow", false, "render drop shadow")
	)
	flag.Parse()

	var input io.Reader

	if flag.NArg() == 0 {
		input = os.Stdin
		*syntaxName = "go"
	} else {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			fmt.Printf("cannot open %q: %+v\n", flag.Arg(0), err)
			os.Exit(1)
		}
		defer f.Close()
		input = f
		*syntaxName = path.Ext(flag.Arg(0))[1:]
	}

	themeData, err := themesFS.Open("themes/" + *themeName + ".json")
	if err != nil {
		fmt.Println("themes.Open", err)
		os.Exit(1)
	}
	var t theme
	if err := json.NewDecoder(themeData).Decode(&t); err != nil {
		fmt.Println("decode theme", err)
		os.Exit(1)
	}

	t.uswWindowDeco = *windowDecorations
	t.useDropShadow = *dropShadow

	if err := renderSnippet(input, *syntaxName, t, *out); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type (
	theme struct {
		Colors        map[string]jsonColor `json:"colors"`
		WindowPadding float64              `json:"windowPadding"`
		CodePadding   float64              `json:"codePadding"`
		BorderRadius  float64              `json:"borderRadius"`
		FontSize      float64              `json:"fontSize"`
		Font          string               `json:"font"`

		useDropShadow bool
		uswWindowDeco bool
	}

	jsonColor struct {
		color color.Color
	}
)

func (c jsonColor) RGBA() (r, g, b, a uint32) {
	if c.color == nil {
		return 0xff, 0x00, 0xff, 0xff
	}
	return c.color.RGBA()
}

func (c *jsonColor) UnmarshalJSON(i []byte) error {
	var err error
	c.color, err = parseColor(string(i)[1 : len(i)-1])
	return err
}

func parseColor(s string) (color.Color, error) {
	c := color.RGBA{A: 0xFF}
	var err error
	switch len(s) {
	case 9:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x%02x", &c.R, &c.G, &c.B, &c.A)
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		return nil, fmt.Errorf("invalid length, must be 9, 7 or 4")
	}
	return c, err
}

func textBounds(buf []byte) (int, int) {
	max := 0
	lines := bytes.Split(buf, []byte("\n"))
	for i := range lines {
		if l := len(lines[i]); l > max {
			max = l
		}
	}
	return max, len(lines)
}

func renderSnippet(contents io.Reader, syntax string, t theme, out string) error {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, contents); err != nil {
		return err
	}

	radius := 7.0
	cWidth, cHeight := textBounds(bytes.TrimSpace(buf.Bytes()))
	width := float64(cWidth)*t.FontSize*1.1 + 2*t.WindowPadding + 2*t.CodePadding
	height := float64(cHeight)*t.FontSize*1.7 + 52 + 2*t.WindowPadding + 2*t.CodePadding

	if !t.uswWindowDeco {
		height -= radius * 3 * 1.7
	}

	var dest interface{}
	var gc draw2d.GraphicContext

	switch path.Ext(out) {
	case ".png":
		dest = image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		gc = draw2dimg.NewGraphicContext(dest.(draw.Image))
	case ".svg":
		svg := draw2dsvg.NewSvg()
		svg.Width = strconv.Itoa(int(math.Ceil(width)))
		svg.Height = strconv.Itoa(int(math.Ceil(height)))
		dest = svg
		gc = draw2dsvg.NewGraphicContext(dest.(*draw2dsvg.Svg))
	default:
		return fmt.Errorf("no supported output format %q", path.Ext(out))
	}

	// Drop shadow hack.
	if t.useDropShadow {
		img := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))
		wc := draw2dimg.NewGraphicContext(img)
		wc.FillStroke()

		wc.Save()
		wc.SetFillColor(t.Colors["background"])
		wc.ClearRect(0, 0, int(width), int(height))
		wc.Fill()

		r, g, b, _ := t.Colors["window"].RGBA()
		if int(r>>8)+int(g>>8)+int(b>>8) > 200*3 {
			r /= 2 << 8
			g /= 2 << 8
			b /= 2 << 8
		}
		blurredColor := color.RGBA{
			uint8(r), uint8(g), uint8(b), 0xFF,
		}
		wc.SetFillColor(blurredColor)
		draw2dkit.RoundedRectangle(
			wc,
			15+t.WindowPadding,
			15+t.WindowPadding,
			20+width-t.WindowPadding,
			20+height-t.WindowPadding,
			t.BorderRadius,
			t.BorderRadius,
		)
		wc.Fill()

		blurred, err := stackblur.Run(img, 50)
		if err != nil {
			return err
		}
		wc.Close()
		gc.DrawImage(blurred)
	} else {
		gc.Save()
		gc.SetFillColor(t.Colors["background"])
		gc.ClearRect(0, 0, int(width), int(height))
		gc.FillStroke()
	}

	// Window.
	{
		gc.Save()
		gc.SetStrokeColor(t.Colors["window"])
		gc.SetFillColor(t.Colors["window"])
		gc.SetLineWidth(1) // border width
		draw2dkit.RoundedRectangle(
			gc,
			t.WindowPadding,
			t.WindowPadding,
			width-t.WindowPadding,
			height-t.WindowPadding,
			t.BorderRadius,
			t.BorderRadius,
		)
		gc.FillStroke()
	}

	lineStartX := t.WindowPadding + t.FontSize*0.5
	lineStartY := t.WindowPadding

	// Buttons.
	if t.uswWindowDeco {
		for i, col := range []color.RGBA{red, yellow, green} {
			gc.Save()
			gc.SetStrokeColor(col)
			gc.SetFillColor(col)
			gc.SetLineWidth(1)
			draw2dkit.Circle(
				gc,
				t.WindowPadding+radius+3.3*radius*float64(i)+2*radius,
				t.WindowPadding+3*radius,
				radius,
			)
			gc.FillStroke()
			lineStartY += radius * 1.7
		}
	}

	//fontData, err := fontFS.ReadFile("fonts/SourceCodePro-Semibold.ttf")
	fontData, err := fontFS.ReadFile("fonts/" + t.Font + ".ttf")
	if err != nil {
		return fmt.Errorf("cannot open font: %+v", err)
	}
	parsedFont, err := truetype.Parse(fontData)
	if err != nil {
		return fmt.Errorf("cannot parse font: %+v", err)
	}
	syntaxData, err := syntaxFS.ReadFile("syntax/" + syntax + ".yaml")
	if err != nil {
		return fmt.Errorf("cannot open syntax: %+v", err)
	}

	// Render text.
	fd := draw2d.FontData{}
	draw2d.RegisterFont(fd, parsedFont)
	gc.SetFontData(fd)
	gc.SetFontSize(t.FontSize)

	syntaxDef, err := highlight.ParseDef(syntaxData)
	if err != nil {
		return fmt.Errorf("syntax highlight: %+v", err)
	}

	buffer := strings.Replace(buf.String(), "\t", "    ", -1)
	buffer = strings.TrimSpace(buffer)
	h := highlight.NewHighlighter(syntaxDef)
	matches := h.HighlightString(buffer)
	for lineN, l := range strings.Split(buffer, "\n") {
		lineStartY += t.FontSize * 1.7

		//if 1+lineN >= t.Highlight.from && 1+lineN <= t.Highlight.to {
		//	gc.Save()
		//	gc.SetFillColor(color.RGBA{0x33, 0x33, 0x37, 0xFF})
		//	draw2dkit.Rectangle(
		//		gc,
		//		t.WindowPadding,
		//		lineStartY-10-fontSize/2,
		//		t.WindowPadding+width-2*t.WindowPadding,
		//		lineStartY+fontSize-6,
		//	)
		//	gc.Fill()
		//}

		gc.SetFillColor(t.Colors["lineNumber"])
		gc.FillStringAt(
			fmt.Sprintf("%3s", strconv.Itoa(lineN+1)),
			lineStartX,
			lineStartY,
		)

		var col color.Color = image.White
		for colN, c := range l {
			if group, ok := matches[lineN][colN]; ok {
				col = t.Colors["foreground"]
				for k := range highlight.Groups {
					if group == highlight.Groups[k] {
						found, ok := t.Colors[k]
						if ok {
							col = found
						}
						break
					}
				}
			}
			gc.SetFillColor(col)
			gc.FillStringAt(
				string(c),
				lineStartX+46+t.FontSize*float64(colN)*0.8,
				lineStartY,
			)
		}

		if group, ok := matches[lineN][len(l)]; ok {
			if group == highlight.Groups["default"] || group == highlight.Groups[""] {
				col = t.Colors["foreground"]
			}
		}
	}

	switch path.Ext(out) {
	case ".png":
		return draw2dimg.SaveToPngFile(out, dest.(draw.Image))
	case ".svg":
		return draw2dsvg.SaveToSvgFile(out, dest.(*draw2dsvg.Svg))
	default:
		return fmt.Errorf("invalid file type %q", path.Ext(out))
	}
}
