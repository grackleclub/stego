package cryptogif

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
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

	// colorPaletteFx := make([]int, 256)

	for _, img := range g.Image {

		compressPalette(img)

		// bounds := img.Bounds()
		// for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		// 	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		// 		// color := img.At(x, y)
		// 		// index := img.Palette.Index(color)
		// 		// colorPaletteFx[uint8(index)]++
		// 		// r8 := uint8(r >> 8)
		// 		// g8 := uint8(g >> 8)
		// 		// b8 := uint8(b >> 8)
		// 		// img.SetColorIndex(x, y, 0)
		// 	}
		// }
	}

	// for i, fx := range colorPaletteFx {
	// 	slog.Debug("color palette fx", "index", i, "fx", fx)
	// }

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

// colorDistance calculates the Euclidean distance between two colors in the RGB color space.
func colorDistance(c1, c2 color.Color) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	// Convert to 8-bit values
	r1, g1, b1 = r1>>8, g1>>8, b1>>8
	r2, g2, b2 = r2>>8, g2>>8, b2>>8

	// Calculate the Euclidean distance
	dr := float64(r1 - r2)
	dg := float64(g1 - g2)
	db := float64(b1 - b2)

	return dr*dr + dg*dg + db*db
}

// findMostSimilarColors finds the 16 colors in the palette that are most similar to other colors.
func findMostSimilarColors(palette color.Palette) []int {
	type colorPair struct {
		index1, index2 int
		distance       float64
	}

	var pairs []colorPair

	// Calculate the distance between each pair of colors
	for i := 0; i < len(palette); i++ {
		for j := i + 1; j < len(palette); j++ {
			distance := colorDistance(palette[i], palette[j])
			pairs = append(pairs, colorPair{i, j, distance})
		}
	}

	// Sort the pairs by distance
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].distance < pairs[j].distance
	})

	// Collect the indices of the 16 most similar colors
	similarColors := make(map[int]struct{})
	for _, pair := range pairs {
		if len(similarColors) >= 16 {
			break
		}
		similarColors[pair.index1] = struct{}{}
		similarColors[pair.index2] = struct{}{}
	}

	// Convert the map keys to a slice
	var indices []int
	for index := range similarColors {
		indices = append(indices, index)
	}

	return indices
}

// compressPalette removes the 16 most similar colors from the palette and reassigns pixels.
func compressPalette(img *image.Paletted) {
	// Find the 16 most similar colors
	similarColors := findMostSimilarColors(img.Palette)

	// Create a new palette without the 16 most similar colors
	newPalette := make(color.Palette, 0, len(img.Palette)-16)
	colorMap := make(map[int]int) // Map old indices to new indices
	for i, c := range img.Palette {
		if !slices.Contains(similarColors, i) {
			colorMap[i] = len(newPalette)
			newPalette = append(newPalette, c)
		}
	}

	// Reassign pixels to the closest remaining colors
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			oldIndex := img.ColorIndexAt(x, y)
			if newIndex, ok := colorMap[int(oldIndex)]; ok {
				img.SetColorIndex(x, y, uint8(newIndex))
			} else {
				// Find the closest remaining color
				_, closestIndex := findMostSimilarColor(newPalette, img.Palette[oldIndex])
				img.SetColorIndex(x, y, uint8(closestIndex))
			}
		}
	}

	// Update the image palette
	img.Palette = newPalette
}

// findMostSimilarColor finds the most similar color in the palette to the target color.
func findMostSimilarColor(palette color.Palette, target color.Color) (color.Color, int) {
	minDistance := float64(1<<32 - 1) // Max float64 value
	closestIndex := 0

	for i, c := range palette {
		if i == 0 || i == 255 {
			continue
		}
		distance := colorDistance(c, target)
		if distance < minDistance {
			minDistance = distance
			closestIndex = i
		}
	}

	return palette[closestIndex], closestIndex
}
