package cryptogif

import (
	"bytes"
	"fmt"
	"image/color"
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

// Read reads the gif at file and returns the gif.GIF.
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

// Encode writes the data to be encoded into the gif at inFile,
// writing the result to outFile.
func Encode(data []byte, inFile, outFile string) error {
	nibbles := toNibbles(data)
	slog.Debug("crushed bytes to nibbles", "nibbles", len(nibbles), "bytes", len(data))

	g, err := Read(inFile)
	if err != nil {
		return fmt.Errorf("read %q: %w", inFile, err)
	}
	slog.Debug("palette may be same ", "true", isCommonPalette(g))

	paletteByIndex, err := newPaletteInfo(g)
	if err != nil {
		return fmt.Errorf("new palette info: %w", err)
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
		return fmt.Errorf("not enough space in gif to encode %d bytes", len(data))
	}

	f, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("create file %q: %w", outFile, err)
	}
	defer f.Close()

	err = gif.EncodeAll(f, g)
	if err != nil {
		return fmt.Errorf("encode gif: %w", err)
	}

	return nil
}

// Decode reads the gif at inFile and returns the embedded data.
func Decode(inFile string) ([]byte, error) {
	slog.Debug("decoding", "file", inFile)
	g, err := Read(inFile)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", inFile, err)
	}

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
	return toBytes(nibbles), nil
}

// isCommonPalette returns true if the palette is the same across all frames.
func isCommonPalette(g *gif.GIF) bool {
	var palettes []color.Palette
	for _, img := range g.Image {
		palettes = append(palettes, img.Palette)
	}
	for i, palette := range palettes {
		for j, p := range palette {
			r, g, b, _ := p.RGBA()
			r0, g0, b0, _ := palettes[0][j].RGBA()
			if r != r0 || g != g0 || b != b0 {
				slog.Debug("palette is inconsistent across frames", "i", i, "j", j, "r", r>>8, "g", g>>8, "b", b>>8)
				return false
			}
		}
	}
	return true
}
