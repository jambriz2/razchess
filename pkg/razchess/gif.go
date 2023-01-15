package razchess

import (
	"image"
	"image/gif"
	"io"

	"github.com/andybons/gogif"
	"github.com/notnil/chess"
	"github.com/razzie/chessimage"
)

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

	return r.Render(chessimage.Options{})
}

func convertImagesToGif(w io.Writer, images []image.Image, delay int) error {
	outGif := &gif.GIF{
		LoopCount: -1,
	}
	for _, img := range images {
		bounds := img.Bounds()
		palettedImage := image.NewPaletted(bounds, nil)
		quantizer := gogif.MedianCutQuantizer{NumColor: 64}
		quantizer.Quantize(palettedImage, bounds, img, image.Point{})
		outGif.Image = append(outGif.Image, palettedImage)
		outGif.Delay = append(outGif.Delay, delay)
	}
	return gif.EncodeAll(w, outGif)
}
