package cryptogif

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"image/color"
	"image/gif"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func init() {
	if testing.Testing() {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
}

func Read(file string) (*gif.GIF, error) {
	if filepath.Ext(file) != ".gif" {
		return nil, fmt.Errorf("file %q is not a gif", file)
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", file, err)
	}
	gif, err := gif.DecodeAll(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("decode gif: %w", err)
	}

	slog.Debug("details",
		"path", file,
		"height", gif.Config.Height,
		"width", gif.Config.Width,
		// "color_model_len", gif.Config.ColorModel,
		"frames", len(gif.Image),
		"background", gif.BackgroundIndex,
		"loop", gif.LoopCount,
	)
	return gif, nil
}

func newSecret(length int) ([]byte, error) {
	secret := make([]byte, length)
	_, err := rand.Read(secret)
	if err != nil {
		return nil, fmt.Errorf("generate secret: %w", err)
	}
	return secret, nil
}

// set floor takes all rgb values below the floor and sets them to floor,
// freeing space for encoding cypertext.
func setFloor(in, out string) ([]byte, error) {
	floor := uint8(64)
	var altered int

	g, err := Read(in)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", in, err)
	}
	for _, img := range g.Image {
		// slog.Debug("frame",
		// 	"index", i,
		// 	"bounds", img.Bounds(),
		// 	"delay", g.Delay[i],
		// 	// "palette", img.Palette,
		// 	// "pixels", img.Pix,
		// )
		bounds := img.Bounds()
		// slog.Debug("frame bounds",
		// 	"x_min", bounds.Min.X,
		// 	"x_max", bounds.Max.X,
		// 	"y_min", bounds.Min.Y,
		// 	"y_max", bounds.Max.Y,
		// )

		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {

				// slog.Debug("pixel",
				// 	"x", x,
				// 	"y", y,
				// 	"color", img.At(x, y),
				// )

				var changed bool
				r, g, b, a := img.At(x, y).RGBA()
				r8 := uint8(r >> 8)
				g8 := uint8(g >> 8)
				b8 := uint8(b >> 8)

				if r8 < floor {
					r8 = floor
					changed = true
				}
				if g8 < floor {
					g8 = floor
					changed = true
				}
				if b8 < floor {
					b8 = floor
					changed = true
				}

				if changed {
					img.Set(x, y, color.RGBA{R: r8, G: g8, B: b8, A: uint8(a >> 8)})
					altered++
				}

				// if i == 10 {
				// 	slog.Debug("pixel",
				// 		"x", x,
				// 		"y", y,
				// 		"color", img.At(x, y),
				// 		"rgba", fmt.Sprintf("%d %d %d %d", r, g, b, a),
				// 		"changed", changed,
				// 	)
				// }

			}
		}
	}

	// write changes to out file

	f, err := os.Create(out)
	if err != nil {
		return nil, fmt.Errorf("create file %q: %w", out, err)
	}
	defer f.Close()

	err = gif.EncodeAll(f, g)
	if err != nil {
		return nil, fmt.Errorf("encode gif: %w", err)
	}

	slog.Debug("altered pixels", "count", altered)

	return nil, nil
}
