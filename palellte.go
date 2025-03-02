package stego

import (
	"fmt"
	"image/color"
	"image/gif"
	"log/slog"
	"sort"
)

// palette info allows determing if the color palette
// should be compressed to allow a floor for data encoding.
type paletteInfo struct {
	color    color.Color // color in palette
	index    uint8       // index in palette
	tone     int         // sum of r, g, b
	toneRank int         // rank based on darkness
}

// newPaletteInfo returns a slice for each frame of the gif, containing
// a slice for every color in the palette, with information about the color,
// including its tone (brightness) and its rank in the palette by tone.
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

// sortByTone sorts a slice of paletteInfo by tone (brightness).
func sortByTone(pis []paletteInfo) []paletteInfo {
	result := make([]paletteInfo, len(pis))
	copy(result, pis)
	sort.Slice(result, func(i, j int) bool {
		return result[i].tone < result[j].tone
	})
	return result
}

// isCommonPalette returns true if the palette is the same across all frames.
func isCommonPalette(g *gif.GIF) bool {
	var palettes []color.Palette
	for _, img := range g.Image {
		palettes = append(palettes, img.Palette)
	}
	for i, palette := range palettes {
		for j, p := range palette {
			r0, g0, b0, _ := palettes[0][j].RGBA()
			r, g, b, _ := p.RGBA()
			if r != r0 || g != g0 || b != b0 {
				slog.Debug(
					"palette found inconsistent with first frame",
					"frame", i,
					"palette", j,
					"r0", r0>>8,
					"g0", g0>>8,
					"b0", b0>>8,
					"r", r>>8,
					"g", g>>8,
					"b", b>>8,
				)
				return false
			}
		}
	}
	return true
}
