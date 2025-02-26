package cryptogif

import (
	"bytes"
	"fmt"
	"image/gif"
	"log/slog"
	"os"
	"path/filepath"
)

// read reads the gif at file and returns a pointer to *gif.GIF.
func read(file string) (*gif.GIF, error) {
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

// write encodes the gif as given, and writes it to the file at path.
func write(g *gif.GIF, path string) error {
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
