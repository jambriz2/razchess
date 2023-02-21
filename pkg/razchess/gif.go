package razchess

import (
	"image"
	"image/color"
	"image/draw"
	"io"

	"github.com/notnil/chess"
	"github.com/razzie/chessimage"
	"github.com/razzie/razchess/pkg/razchess/internal"
)

const boardSize = 512

var palette = getPalette()

func MoveHistoryToGIF(w io.Writer, moves []*chess.Move, positions []*chess.Position) error {
	renderers := make([]*chessimage.Renderer, 0, len(positions))
	initialPos := positions[0]
	positions = positions[1:]

	img, err := prepareMoveRenderer(initialPos, nil)
	if err != nil {
		return err
	}
	renderers = append(renderers, img)

	for i, move := range moves {
		img, err := prepareMoveRenderer(positions[i], move)
		if err != nil {
			return err
		}
		renderers = append(renderers, img)
	}

	images := make(chan *image.Paletted)
	go func() {
		for _, r := range renderers {
			img, _ := r.Render(chessimage.Options{
				PieceRatio: 1,
				BoardSize:  boardSize,
			})
			bounds := img.Bounds()
			palettedImage := image.NewPaletted(bounds, palette)
			draw.Draw(palettedImage, bounds, img, image.Point{}, draw.Over)
			images <- palettedImage
		}
		close(images)
	}()

	return convertImagesToGif(w, images, 100)
}

func prepareMoveRenderer(pos *chess.Position, move *chess.Move) (*chessimage.Renderer, error) {
	r, err := chessimage.NewRendererFromFEN(pos.String())
	if err != nil {
		return nil, err
	}

	if move != nil {
		from, _ := chessimage.TileFromAN(move.S1().String())
		to, _ := chessimage.TileFromAN(move.S2().String())
		r.SetLastMove(chessimage.LastMove{
			From: from,
			To:   to,
		})
		if move.HasTag(chess.Check) {
			kingSq, _ := chessimage.TileFromAN(pos.Board().KingSquare(pos.Turn()).String())
			r.SetCheckTile(kingSq)
		}
	}

	return r, nil
}

func convertImagesToGif(w io.Writer, images <-chan *image.Paletted, delay int) error {
	return internal.Encode(w, image.Point{X: boardSize, Y: boardSize}, images, delay, -1)
}

func rgb(r, g, b uint8) color.Color {
	return &color.RGBA{R: r, G: g, B: b, A: 255}
}

func mix(c1, c2 color.Color) color.Color {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	return &color.RGBA{
		R: uint8((r1 + r2) / 2),
		G: uint8((g1 + g2) / 2),
		B: uint8((b1 + b2) / 2),
		A: 255,
	}
}

func getPalette() []color.Color {
	lightSq := rgb(240, 217, 181)
	darkSq := rgb(181, 136, 99)
	lightSqHigh := rgb(247, 193, 99)
	darkSqHigh := rgb(215, 149, 54)
	check := rgb(255, 0, 0)

	var palette []color.Color
	pieceColors := []color.Color{color.White, color.Black, &color.Gray{Y: 128}}
	sqColors := []color.Color{lightSq, darkSq, lightSqHigh, darkSqHigh, check}

	palette = append(palette, pieceColors...)
	palette = append(palette, sqColors...)
	for _, pieceColor := range pieceColors {
		for _, sqColor := range sqColors {
			palette = append(palette, mix(pieceColor, sqColor))
		}
	}
	return palette
}
