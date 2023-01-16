package razchess

import (
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"io"

	"github.com/notnil/chess"
	"github.com/razzie/chessimage"
)

var palette = getPalette()

func MoveHistoryToGIF(w io.Writer, moves []*chess.Move, positions []*chess.Position) error {
	var images []image.Image

	initialPos := positions[0]
	positions = positions[1:]

	img, err := moveToImage(initialPos, nil)
	if err != nil {
		return err
	}
	images = append(images, img)

	for i, move := range moves {
		img, err := moveToImage(positions[i], move)
		if err != nil {
			return err
		}
		images = append(images, img)
	}

	return convertImagesToGif(w, images, 100)
}

func moveToImage(pos *chess.Position, move *chess.Move) (image.Image, error) {
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
			var kingSq chessimage.Tile
			if pos.Turn() == chess.White {
				kingSq, _ = chessimage.TileFromAN(pos.Board().WhiteKingSquare().String())
			} else {
				kingSq, _ = chessimage.TileFromAN(pos.Board().BlackKingSquare().String())
			}
			r.SetCheckTile(kingSq)
		}
	}

	return r.Render(chessimage.Options{
		PieceRatio: 1,
		BoardSize:  512,
	})
}

func convertImagesToGif(w io.Writer, images []image.Image, delay int) error {
	outGif := &gif.GIF{
		LoopCount: -1,
	}
	for _, img := range images {
		bounds := img.Bounds()
		palettedImage := image.NewPaletted(bounds, palette)
		draw.Draw(palettedImage, bounds, img, image.Point{}, draw.Over)
		outGif.Image = append(outGif.Image, palettedImage)
		outGif.Delay = append(outGif.Delay, delay)
	}
	return gif.EncodeAll(w, outGif)
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
