package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	_ "image/gif"
	_ "image/png"
)

var errHomeNotFound = errors.New("wimg: user home directory not found")

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
	home := getUserHome()
	if home == "" {
		return errHomeNotFound
	}
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
	return save(home, name, resp.Body)
}

func getUserHome() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
	return home
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

func save(home, name string, r io.Reader) error {
	m, _, err := image.Decode(r)
	if err != nil {
		return err
	}
	rect := m.Bounds()
	filename := getSaveName(home, name, rect)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, m, nil)
}

func getSaveName(home, name string, rect image.Rectangle) string {
	name = fmt.Sprintf("%dx%d_%s.jpg", rect.Dx(), rect.Dy(), name)
	return filepath.Join(home, "img", name)
}
