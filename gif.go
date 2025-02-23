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

func init() {
	if testing.Testing() {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}
}

type nibblePalette struct {
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
	slog.Info("common palette", "true", isCommonPalette(g))

	paletteByIndex, err := newPI(g)
	if err != nil {
		return nil, fmt.Errorf("new palette info: %w", err)
	}

	var currentNibble int
	lastNibble := len(nibbles) - 1
	for i, img := range g.Image {
		bounds := img.Bounds()
		pal := img.Palette
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				// replace dark colors with nibbles of data
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
						slog.Debug("writing nibble", "n", n, "frame", i, "x", x, "y", y)
						newDataColor := paletteByTone[n]
						img.Set(x, y, newDataColor.color)
					}
				}
			}
		}
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

	return nil, nil
}

func decode(inFile string) ([]byte, error) {
	slog.Debug("decoding", "file", inFile)
	g, err := Read(inFile)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", inFile, err)
	}

	paletteByIndex, err := newPI(g)
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
					nibbles = append(nibbles, uint8(p.toneRank))
				}

				if x < 100 && y == 0 {
					slog.Debug("debug", "index", index, "frame", i, "x", x, "y", y)
				}

				// if index < 16 {
				// 	slog.Debug("index", "index", index, "frame", i, "x", x, "y", y)
				// 	nibbles = append(nibbles, uint8(index))

				// 	// TODO temp hard limit
				// 	if len(nibbles) > 64 {
				// 		return toBytes(nibbles), nil
				// 	}
				// }
			}
		}
	}
	return toBytes(nibbles), nil
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

	// TODO fix this hack
	if len(nibbles)%2 != 0 {
		slog.Warn("odd number of nibbles, dropping last", "len", len(nibbles))
		nibbles = nibbles[:len(nibbles)-1]
	}

	for i := 0; i < len(nibbles); i += 2 {
		if len(nibbles) < i+1 {
			stretched = append(stretched, nibbles[i])
			return stretched
		} else {
			stretched = append(stretched, joinNibbles(nibbles[i], nibbles[i+1]))
		}
	}
	return stretched
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
				slog.Debug("palette mismatch", "i", i, "j", j, "r", r>>8, "g", g>>8, "b", b>>8)
				return false
			}
		}
	}
	return true
}

// palette info
type pi struct {
	color    color.Color // color in palette
	index    uint8       // index in palette
	tone     int         // sum of r, g, b
	toneRank int         // rank based on darkness
}

func newPI(g *gif.GIF) ([][]pi, error) {
	assumedLen := 256
	var pisResults [][]pi = make([][]pi, 0)
	for i, img := range g.Image {
		if len(img.Palette) != assumedLen {
			return nil, fmt.Errorf("palette length: got %d, expected %v", len(img.Palette), assumedLen)
		}
		pis := make([]pi, assumedLen)
		for i, color := range img.Palette {
			r, g, b, _ := color.RGBA()
			sum := r>>8 + g>>8 + b>>8
			pis[i].tone = int(sum)
			pis[i].color = color
			pis[i].index = uint8(i)
		}

		ranking := make([]pi, len(pis))
		copy(ranking, pis)
		sort.Slice(ranking, func(i, j int) bool {
			return ranking[i].tone < ranking[j].tone
		})

		for rank, cpy := range ranking {
			pis[cpy.index].toneRank = rank
		}
		slog.Debug("palette info being added", "len", len(pis), "frame", i)
		pisResults = append(pisResults, pis)
	}
	slog.Debug("palette info", "len", len(pisResults))
	return pisResults, nil
}

func sortByTone(pis []pi) []pi {
	result := make([]pi, len(pis))
	copy(result, pis)
	sort.Slice(result, func(i, j int) bool {
		return result[i].tone < result[j].tone
	})
	return result
}
