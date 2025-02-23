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

	pi := newPI(g)
	slog.Debug("palette info", "len", len(pi))

	for _, img := range g.Image {
		bounds := img.Bounds()
		pal := img.Palette
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				index := pal.Index(img.At(x, y))
				p := pi[index]
				if p.toneRank > 250 {
					img.Set(x, y, pi[0].color)
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
	color    color.Color
	index    uint8
	tone     int
	toneRank int
}

func newPI(g *gif.GIF) []pi {
	pis := make([]pi, 256)
	for i, color := range g.Image[0].Palette {
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

	// Assign toneRank based on sorted order
	for rank, cpy := range ranking {
		pis[cpy.index].toneRank = rank
	}

	return pis
}

func tone(g *gif.GIF) []int {
	tone := make([]int, 256)
	for i, color := range g.Image[0].Palette {
		r, g, b, _ := color.RGBA()
		sum := r>>8 + g>>8 + b>>8
		tone[i] = int(sum)
	}
	return tone
}
