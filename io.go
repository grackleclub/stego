package cryptogif

import (
	"bytes"
	"fmt"
	"image/gif"
	"log/slog"
	"os"
)

// Read reads the gif at file and returns a pointer to *gif.GIF.
func Read(filepath string) (*gif.GIF, error) {
	b, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", filepath, err)
	}
	gif, err := gif.DecodeAll(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("decode gif: %w", err)
	}
	info, err := os.Stat(filepath)
	if err != nil {
		return nil, fmt.Errorf("stat file %q: %w", filepath, err)
	}
	slog.Debug("file read",
		"path", filepath,
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
func Write(g *gif.GIF, filepath string) error {
	f, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("create file %q: %w", filepath, err)
	}
	defer f.Close()

	err = gif.EncodeAll(f, g)
	if err != nil {
		return fmt.Errorf("encode gif: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat file %q: %w", filepath, err)
	}
	slog.Debug("file written", "path", filepath, "size", info.Size())
	return nil
}
