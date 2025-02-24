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

// The 'floor' is the central concept which drives encoding of data.
// Of the up to 256 colors in each gif frame's palette, the darkest colors,
// defined as ranked in tone below floor, are altered (assigned to the palette at floor+1),
// leaving space below the floor for encoding data.
var floor = 16

func init() {
	if testing.Testing() {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
}

// Read reads the gif at file and returns a pointer to *gif.GIF.
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
	info, err := os.Stat(file)
	if err != nil {
		return nil, fmt.Errorf("stat file %q: %w", file, err)
	}
	slog.Debug("file read",
		"path", file,
		"height", gif.Config.Height,
		"width", gif.Config.Width,
		"frames", len(gif.Image),
		"background", gif.BackgroundIndex,
		"loop", gif.LoopCount,
		"size", info.Size(),
	)
	return gif, nil
}

// Write encodes the gif as given, and writes it to the file at path.
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
	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat file %q: %w", path, err)
	}
	slog.Debug("file written", "path", path, "size", info.Size())
	return nil
}

// Encode first converts data to base64, then to nibbles,
// altering the gif to provide a "floor" for cyphertext,
// inserting data one nibble at a time until completion,
// or exhaustion of the gif's capacity.
func Encode(g *gif.GIF, data []byte) (*gif.GIF, error) {
	b64 := base64.StdEncoding.EncodeToString(data)
	slog.Debug("encoding to base64", "data", b64)

	nibbles, err := toNibbles([]byte(b64))
	if err != nil {
		return nil, fmt.Errorf("to nibbles: %w", err)
	}
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
				if currentNibble > lastNibble {
					continue
				}
				if currentNibble == lastNibble {
					slog.Debug("writing last nibble", "value", nibbles[currentNibble], "frame", i, "x", x, "y", y)
					newDataColor := paletteByTone[floor]
					img.Set(x, y, newDataColor.color)
					currentNibble++
				}
				if currentNibble <= lastNibble && p.toneRank < floor {
					n := nibbles[currentNibble]
					// slog.Debug("writing nibble", "value", n, "frame", i, "x", x, "y", y)
					newDataColor := paletteByTone[n]
					img.Set(x, y, newDataColor.color)
					currentNibble++
					continue
				}
			}
		}
	}
	if currentNibble < lastNibble {
		return nil, fmt.Errorf("not enough space in gif to encode %d bytes", len(data))
	}
	return g, nil
}

// Decode reads the gif and decodes the data expecting the same as Encode:
//   - base64 encoded data
//   - crushed to nibbles
//   - inserted into the gif at image palette[0] through palette[floor]
func Decode(g *gif.GIF) ([]byte, error) {
	slog.Info("decoding gif")
	paletteByIndex, err := newPaletteInfo(g)
	if err != nil {
		return nil, fmt.Errorf("new palette info: %w", err)
	}

	var nibbles []uint8
extraction:
	for i, img := range g.Image {
		bounds := img.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				index := img.Palette.Index(img.At(x, y))
				p := paletteByIndex[i][index]
				if p.toneRank == floor {
					break extraction
				}
				if p.toneRank <= floor {
					// slog.Debug("reading nibble", "value", p.toneRank, "frame", i, "x", x, "y", y)
					nibbles = append(nibbles, uint8(p.toneRank))
				}
			}
		}
	}
	bytes, err := toBytes(nibbles)
	if err != nil {
		return nil, fmt.Errorf("to bytes: %w", err)
	}
	result, err := base64.StdEncoding.DecodeString(string(bytes))
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}
	return result, nil
}
