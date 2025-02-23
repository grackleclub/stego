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

var floor = uint8(16) // the minimum value for r, g, b to set a 'floor', below which is reserved

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
func encode(data []byte, inFile, outFile string) ([]byte, error) {
	nibbles := toNibbles(data)
	slog.Debug("crushed bytes to nibbles", "nibbles", len(nibbles), "bytes", len(data))

	g, err := Read(inFile)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", inFile, err)
	}

	currentNibble := 0
	lastNibble := len(nibbles)
	var done bool
	for i, img := range g.Image {
		if done {
			slog.Debug("done writing", "frame", i)
			break
		}
		bounds := img.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, a := img.At(x, y).RGBA()
				r8 := uint8(r >> 8)
				g8 := uint8(g >> 8)
				b8 := uint8(b >> 8)

				var alter bool
				// pixel color is below floor and eligible for placing data
				if r8 <= floor {
					alter = true
					// write the next nibble
					if currentNibble < lastNibble {
						r8 = nibbles[currentNibble]
						currentNibble++
						slog.Info("wriring data", "r8", r8, "currentNibble", currentNibble, "frame", i)
					} else if currentNibble == lastNibble {
						// write the last byte at the floor
						r8 = floor
						currentNibble++
						slog.Warn("writing end marker", "r8", r8, "currentNibble", currentNibble, "frame", i)
						done = true
					}
				}

				if alter {
					img.Set(x, y, color.RGBA{R: r8, G: g8, B: b8, A: uint8(a >> 8)})
				}
			}
		}
	}

	if currentNibble < len(nibbles) {
		// TODO be detailed
		return nil, fmt.Errorf("not enough space")
	}

	// write changes to out file
	f, err := os.Create(outFile)
	if err != nil {
		return nil, fmt.Errorf("create file %q: %w", outFile, err)
	}
	defer f.Close()

	err = gif.EncodeAll(f, g)
	if err != nil {
		return nil, fmt.Errorf("encode gif: %w", err)
	}

	slog.Debug("data written", "nibbles", currentNibble) // don't include end marker

	return nil, nil
}

func decode(inFile string) ([]byte, error) {
	slog.Debug("decoding", "file", inFile)
	g, err := Read(inFile)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", inFile, err)
	}

	// var nibbles []uint8
	for i, img := range g.Image {
		bounds := img.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, _, _, _ := img.At(x, y).RGBA()
				r8 := uint8(r >> 8)
				if i == 0 && y == 0 && r8 <= floor {
					slog.Debug("found value", "r8", r8, "floor", floor, "frame", i, "x", x, "y", y)
				}

				// if r8 == floor {
				// 	slog.Debug("end marker reached", "frame", i, "x", x, "y", y)
				// 	return toBytes(nibbles), nil
				// }
				// if r8 < floor {
				// 	slog.Warn("value detected", "r8", r8, "frame", i, "x", x, "y", y)
				// 	nibbles = append(nibbles, r8)
				// }
			}
		}
	}
	return nil, fmt.Errorf("no end marker found")
}

func splitNibbles(b byte) (byte, byte) {
	return b >> 4, b & 0x0f
}

func joinNibbles(n1, n2 byte) byte {
	return n1<<4 | n2
}

// toNibbles takes a byte slice and returns a slice of nibbles,
// guaranteed to be even in length.
func toNibbles(bytes []byte) []uint8 {
	var crushed []uint8
	for _, b := range bytes {
		n1, n2 := splitNibbles(b)
		crushed = append(crushed, n1, n2)
	}
	slog.Info("crushed", "len_nib", len(crushed), "len_bytes", len(bytes))
	return crushed
}

// toBytes takes an even length slice of nibbles and returns a byte slice.
func toBytes(nibbles []uint8) []byte {
	var stretched []byte

	if len(nibbles)%2 != 0 {
		slog.Warn("odd number of nibbles, dropping last", "len", len(nibbles))
		nibbles = nibbles[:len(nibbles)-1]
	}

	for i := 0; i < len(nibbles); i += 2 {
		slog.Info("stretched", "i", i, "len", len(nibbles))
		if len(nibbles) < i+1 {
			stretched = append(stretched, nibbles[i])
			return stretched
		} else {
			stretched = append(stretched, joinNibbles(nibbles[i], nibbles[i+1]))
		}
	}
	return stretched
}
