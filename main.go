package main

import (
	"flag"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	dir     = flag.String("dir", "", "directory to save file (default: cwd)")
	name    = flag.String("name", "", "name of the image")
	format  = flag.String("format", "jpg", "format to save image (default: jpg)")
	quality = flag.Int("quality", 100, "jpeg quality encoding")
)

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "wimg: invalid arguments")
		flag.PrintDefaults()
		os.Exit(1)
	}
	err := run(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(src string) error {
	err := isSupportedFormat(*format)
	if err != nil {
		return err
	}
	name, err := normalize(*name, src)
	if err != nil {
		return err
	}
	resp, err := http.Get(src)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return save(resp.Body, *dir, name, *format, *quality)
}

func isSupportedFormat(format string) error {
	var formats = []string{"gif", "jpg", "png"}
	for _, f := range formats {
		if f == format {
			return nil
		}
	}
	return fmt.Errorf("Format must be jpg (default), png or gif.")
}

func normalize(name, src string) (string, error) {
	if name == "" {
		name = baseWithoutExt(src)
	}
	t := transform.Chain(norm.NFD, transform.RemoveFunc(remove), norm.NFC)
	name = strings.TrimSpace(name)
	name, _, err := transform.String(t, name)
	if err != nil {
		return "", err
	}
	name = strings.ToLower(name)
	name = strings.Replace(name, " ", "_", -1)
	return name, nil
}

func baseWithoutExt(src string) string {
	ext := path.Ext(src)
	name := path.Base(src)
	return strings.TrimSuffix(name, ext)
}

func remove(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}

func save(r io.Reader, dir, name, format string, quality int) error {
	m, _, err := image.Decode(r)
	if err != nil {
		return err
	}
	rect := m.Bounds()
	filename := getSaveName(rect, dir, name, format)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	switch format {
	case "gif":
		err = gif.Encode(f, m, nil)
	case "jpg":
		err = jpeg.Encode(f, m, &jpeg.Options{Quality: quality})
	case "png":
		err = png.Encode(f, m)
	}
	return err
}

func getSaveName(rect image.Rectangle, dir, name, format string) string {
	filename := fmt.Sprintf("%dx%d_%s.%s", rect.Dx(), rect.Dy(), name, format)
	return filepath.Join(dir, filename)
}
