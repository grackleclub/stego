package cryptogif

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/gif"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

var floor = 16 // the minimum value for r, g, b to set a 'floor', below which is reserved

func init() {
	if testing.Testing() {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
}

// Read reads the gif at file and returns a pointer *gif.GIF.
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

	slog.Debug("file read",
		"path", file,
		"height", gif.Config.Height,
		"width", gif.Config.Width,
		"frames", len(gif.Image),
		"background", gif.BackgroundIndex,
		"loop", gif.LoopCount,
	)
	return gif, nil
}

func Encode(g *gif.GIF, data []byte) (*gif.GIF, error) {
	b64 := base64.StdEncoding.EncodeToString(data)
	slog.Debug("encoding to base64", "data", b64)

	nibbles := toNibbles([]byte(b64))
	slog.Debug("crushed bytes to nibbles", "nibbles", len(nibbles), "bytes", len(data))
	slog.Debug("palette homogeneity", "is_same", isCommonPalette(g))

	paletteByIndex, err := newPaletteInfo(g)
	if err != nil {
		return nil, fmt.Errorf("new palette info: %w", err)
	}

	var currentNibble int
	lastNibble := len(nibbles) - 1
	// frame
	for i, img := range g.Image {
		bounds := img.Bounds()
		pal := img.Palette
		// row
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			// column (pixel)
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				// get the index of the pixel's color in the palette
				index := pal.Index(img.At(x, y))
				p := paletteByIndex[i][index]
				paletteByTone := sortByTone(paletteByIndex[i])
				if p.toneRank <= floor {
					if currentNibble > lastNibble {
						// backfill with floor
						img.Set(x, y, paletteByTone[floor+1].color)
					} else {
						// or write the next nibble
						n := nibbles[currentNibble]
						currentNibble++
						slog.Debug("writing nibble", "value", n, "frame", i, "x", x, "y", y)
						newDataColor := paletteByTone[n]
						img.Set(x, y, newDataColor.color)
					}
				}
			}
		}
	}
	if currentNibble <= lastNibble {
		return nil, fmt.Errorf("not enough space in gif to encode %d bytes", len(data))
	}
	return g, nil
}

// Write encodes the gif and writes it to the file at path.
func Write(g *gif.GIF, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file %q: %w", path, err)
	}
	defer f.Close()

	err = gif.EncodeAll(f, g)
	if err != nil {
		return fmt.Errorf("encode gif: %w", err)
	}
	return nil
}

func Decode(g *gif.GIF) ([]byte, error) {
	paletteByIndex, err := newPaletteInfo(g)
	if err != nil {
		return nil, fmt.Errorf("new palette info: %w", err)
	}

	var nibbles []uint8
	for i, img := range g.Image {
		bounds := img.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				index := img.Palette.Index(img.At(x, y))
				p := paletteByIndex[i][index]
				if p.toneRank <= floor {
					slog.Debug("reading nibble", "value", p.toneRank, "frame", i, "x", x, "y", y)
					nibbles = append(nibbles, uint8(p.toneRank))
				}
			}
		}
	}
	bytes, err := base64.StdEncoding.DecodeString(string(toBytes(nibbles)))
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}
	return bytes, nil
}
