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
	"sort"
	"testing"
)

var floor = uint8(16) // the minimum value for r, g, b to set a 'floor', below which is reserved

// palette info
type paletteInfo struct {
	color    color.Color // color in palette
	index    uint8       // index in palette
	tone     int         // sum of r, g, b
	toneRank int         // rank based on darkness
}

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

// set floor takes all rgb values below the floor and sets them to floor,
// freeing space for encoding cypertext.
func encode(data []byte, inFile, outFile string) error {
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
				if p.toneRank <= 16 {
					if currentNibble > lastNibble {
						// backfill with floor
						img.Set(x, y, paletteByTone[17].color)
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

func decode(inFile string) ([]byte, error) {
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
				if len(nibbles) > 128 {
					return toBytes(nibbles), nil
				}
				index := img.Palette.Index(img.At(x, y))
				p := paletteByIndex[i][index]
				if p.toneRank <= 16 {
					slog.Debug("reading nibble", "value", p.toneRank, "frame", i, "x", x, "y", y)
					nibbles = append(nibbles, uint8(p.toneRank))
				}
			}
		}
	}
	return toBytes(nibbles), nil
}

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

func newPaletteInfo(g *gif.GIF) ([][]paletteInfo, error) {
	assumedLen := 256
	var pisResults [][]paletteInfo = make([][]paletteInfo, 0)
	for _, img := range g.Image {
		if len(img.Palette) != assumedLen {
			return nil, fmt.Errorf("palette length: got %d, expected %v", len(img.Palette), assumedLen)
		}
		pis := make([]paletteInfo, assumedLen)
		for i, color := range img.Palette {
			r, g, b, _ := color.RGBA()
			sum := r>>8 + g>>8 + b>>8
			pis[i].tone = int(sum)
			pis[i].color = color
			pis[i].index = uint8(i)
		}

		ranking := make([]paletteInfo, len(pis))
		copy(ranking, pis)
		sort.Slice(ranking, func(i, j int) bool {
			return ranking[i].tone < ranking[j].tone
		})

		for rank, cpy := range ranking {
			pis[cpy.index].toneRank = rank
		}
		pisResults = append(pisResults, pis)
	}
	return pisResults, nil
}

func sortByTone(pis []paletteInfo) []paletteInfo {
	result := make([]paletteInfo, len(pis))
	copy(result, pis)
	sort.Slice(result, func(i, j int) bool {
		return result[i].tone < result[j].tone
	})
	return result
}

func newSecret(length int) ([]byte, error) {
	secret := make([]byte, length)
	_, err := rand.Read(secret)
	if err != nil {
		return nil, fmt.Errorf("generate secret: %w", err)
	}
	return secret, nil
}
