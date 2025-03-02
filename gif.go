package stego

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image/gif"
	"log/slog"
	"strconv"
	"testing"
)

// The 'floor' is the central concept which drives encoding of data.
// Of the up to 256 colors in each gif frame's palette, the darkest colors,
// defined as ranked in tone below floor, are altered (assigned to the palette at floor+1),
// leaving space below the floor for encoding data.
const floor = 17

func init() {
	if testing.Testing() {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
}

// Inject first converts data to base64, then to nibbles,
// altering the gif to provide a "floor" for cyphertext,
// inserting data one nibble at a time until completion,
// or exhaustion of the gif's capacity.
func Inject(g *gif.GIF, data []byte) (*gif.GIF, error) {
	b64 := base64.StdEncoding.EncodeToString(data)
	hex := hex.EncodeToString([]byte(b64))
	slog.Debug("palette homogeneity", "is_same", isCommonPalette(g))

	paletteByIndex, err := newPaletteInfo(g)
	if err != nil {
		return nil, fmt.Errorf("new palette info: %w", err)
	}

	var currentChar int
	footerChar := len(hex)
encoding:
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
				if p.toneRank < floor {
					if currentChar == footerChar {
						slog.Debug("writing floor", "frame", i, "x", x, "y", y)
						newDataColor := paletteByTone[floor]
						img.Set(x, y, newDataColor.color)
						currentChar++
						break encoding
					} else if currentChar < footerChar {
						n := hex[currentChar]
						dec, err := strconv.ParseInt(string(n), 16, 8)
						if err != nil {
							return nil, fmt.Errorf("parse int: %w", err)
						}
						// slog.Debug("writing hex", "value", dec, "frame", i, "x", x, "y", y)
						newDataColor := paletteByTone[dec+1]
						img.Set(x, y, newDataColor.color)
						currentChar++
					}
				}
			}
		}
	}
	// slog.Debug("encoded result", "data", char, "current", currentChar, "footer", footerChar)
	if currentChar <= footerChar {
		return nil, fmt.Errorf("not enough space in gif to encode %d bytes", len(data))
	}
	return g, nil
}

// Extract reads the gif and decodes the data expecting the same as Encode:
//   - inserted into the gif at image palette[1] through palette[floor]
//
// Note: palette[0] is often reserved for background)
func Extract(g *gif.GIF) ([]byte, error) {
	slog.Debug("decoding gif")
	paletteByIndex, err := newPaletteInfo(g)
	if err != nil {
		return nil, fmt.Errorf("new palette info: %w", err)
	}

	var char string
extraction:
	for i, img := range g.Image {
		bounds := img.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				index := img.Palette.Index(img.At(x, y))
				p := paletteByIndex[i][index]
				if p.toneRank == floor {
					slog.Debug("found floor", "frame", i, "x", x, "y", y)
					break extraction
				}
				if p.toneRank < floor {
					char += fmt.Sprintf("%x", p.toneRank-1)
				}
			}
		}
	}
	bytes, err := hex.DecodeString(char)
	if err != nil {
		slog.Info("exiting with partial result", "bytes_extracted", len(bytes))
		return nil, fmt.Errorf("decode string: %w", err)
	}
	decoded, err := base64.StdEncoding.DecodeString(string(bytes))
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}
	return decoded, nil
}
