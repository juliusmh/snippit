package main

import (
	"bytes"
	"fmt"
	"image/color"
	"os"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/golang/freetype/truetype"
	"github.com/urfave/cli/v2"

	"github.com/juliusmh/snippit/draw"
	"github.com/juliusmh/snippit/elements"
)

func main() {
	app := cli.App{
		Name:        "snippit",
		Description: "Create images of code snippets!",
		UsageText:   "snippit <args> [FILE(s)]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "style",
				Value: "rainbow_dash",
				Usage: "Style to use for rendering",
			},
			&cli.StringFlag{
				Name:  "syntax",
				Value: "auto",
				Usage: "Syntax to use for syntax highlighting",
			},
			&cli.Float64Flag{
				Name:  "scale",
				Value: 1,
				Usage: "Scale the rendered image",
			},
			&cli.Float64Flag{
				Name:    "border-radius",
				Value:   7,
				Aliases: []string{"br"},
				Usage:   "Border radius",
			},
			&cli.Float64Flag{
				Name:    "border-width",
				Value:   1.5,
				Aliases: []string{"bw"},
				Usage:   "Border width (0 to disable border)",
			},
			&cli.BoolFlag{
				Name:    "window",
				Value:   false,
				Aliases: []string{"w"},
				Usage:   "Show window decorations",
			},
			&cli.BoolFlag{
				Name:    "lines",
				Value:   false,
				Aliases: []string{"l"},
				Usage:   "Show line numbers",
			},
		},
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				return fmt.Errorf("no file(s) specified")
			}
			style := styles.Get(c.String("style"))
			if style == nil || style == styles.Fallback {
				return fmt.Errorf("unknown style")
			}

			backgroundColor := style.Get(chroma.Background).Background
			linesColor := style.Get(chroma.LineNumbers).Colour

			for _, fileName := range c.Args().Slice() {
				content, lexer, err := getContentAndLexer(c.String("syntax"), fileName)
				if err != nil {
					return err
				}
				padding := 13 * c.Float64("scale")

				column := elements.Column{
					elements.Code{
						Text:  content,
						Lexer: lexer,
						Style: style,
						Face: truetype.NewFace(elements.RobotoMono, &truetype.Options{
							Size: 17 * c.Float64("scale"),
						}),
						ShowLineNumbers: c.Bool("lines"),
						LineNumberColor: convertColor(linesColor),
					},
				}

				if c.Bool("window") {
					column = append([]elements.Element{
						elements.Padding{
							Bottom: padding / 1.5,
							Element: elements.Row{
								elements.Dots{
									Radius:  6 * c.Float64("scale"),
									Spacing: 6 * c.Float64("scale"),
								},
							},
						},
					}, column...)
				}

				drawable := elements.Padding{
					Top:    padding,
					Left:   padding,
					Right:  padding,
					Bottom: padding,
					Element: elements.Card{
						BackgroundColor: convertColor(backgroundColor),
						BorderColor:     darken(convertColor(backgroundColor)),
						BorderWidth:     c.Float64("border-width") * c.Float64("scale"),
						BorderRadius:    c.Float64("border-radius") * c.Float64("scale"),
						Body: elements.Padding{
							Left:    padding,
							Right:   padding,
							Top:     padding,
							Bottom:  padding,
							Element: column,
						},
					},
				}

				w, h := drawable.Measure()
				ctx := draw.GG(w, h, c.Float64("scale"))
				drawable.Draw(ctx, 0, 0)
				outFile := fileName + ".png"
				if err := ctx.SavePNG(outFile); err != nil {
					return err
				}
			}
			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("error: %+v\n", err)
		os.Exit(1)
	}
}

func getContentAndLexer(syntax, fileName string) (string, chroma.Lexer, error) {
	contents, err := os.ReadFile(fileName)
	if err != nil {
		return "", nil, err
	}
	// replace tabs with spaces:
	contents = bytes.Replace(contents, []byte("\t"), []byte("    "), -1)
	var lexer chroma.Lexer
	if syntax == "auto" {
		lexer = lexers.Match(fileName)
		if lexer == nil || lexer == lexers.Fallback {
			lexer = lexers.Analyse(string(contents))
		}
	} else {
		lexer = lexers.Get(syntax)
	}
	if lexer == nil || lexer == lexers.Fallback {
		return "", nil, fmt.Errorf("could not determine lexer")
	}
	return string(contents), lexer, nil
}

func convertColor(col chroma.Colour) color.Color {
	return color.RGBA{
		R: col.Red(),
		G: col.Green(),
		B: col.Blue(),
		A: 0xFF,
	}
}

func darken(col color.Color) color.Color {
	r, g, b, _ := col.RGBA()
	return color.RGBA{
		R: uint8(float64(r) * 0.7 / 255),
		G: uint8(float64(g) * 0.7 / 255),
		B: uint8(float64(b) * 0.7 / 255),
		A: 0xFF,
	}
}
