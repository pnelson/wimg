package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	_ "image/gif"
	_ "image/png"
)

func main() {
	var (
		name    = flag.String("name", "", "name of the image")
		quality = flag.Int("quality", 100, "jpeg quality encoding")
	)
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "wimg: invalid arguments")
		flag.PrintDefaults()
		os.Exit(1)
	}
	err := run(args[0], *name, *quality)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(src, name string, quality int) error {
	if name == "" {
		name = baseWithoutExt(src)
	}
	name, err := normalize(name)
	if err != nil {
		return err
	}
	resp, err := http.Get(src)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return save(name, resp.Body, quality)
}

func baseWithoutExt(src string) string {
	ext := path.Ext(src)
	name := path.Base(src)
	return strings.TrimSuffix(name, ext)
}

func normalize(name string) (string, error) {
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

func remove(r rune) bool {
	return unicode.Is(unicode.Mn, r)
}

func save(name string, r io.Reader, quality int) error {
	m, _, err := image.Decode(r)
	if err != nil {
		return err
	}
	rect := m.Bounds()
	filename := getSaveName(name, rect)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, m, &jpeg.Options{Quality: quality})
}

func getSaveName(name string, rect image.Rectangle) string {
	return fmt.Sprintf("%dx%d_%s.jpg", rect.Dx(), rect.Dy(), name)
}
