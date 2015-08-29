package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
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

func getUserHome() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
	return home
}

func remove(r rune) bool {
	return unicode.Is(unicode.Mn, r)
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

func run(src, name string, quality int) error {
	home := getUserHome()
	if home == "" {
		return errHomeNotFound
	}
	resp, err := http.Get(src)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	m, _, err := image.Decode(resp.Body)
	if err != nil {
		return err
	}
	bounds := m.Bounds()
	if name == "" {
		ext := path.Ext(src)
		name = path.Base(src)
		name = strings.TrimSuffix(name, ext)
	}
	name, err = normalize(name)
	if err != nil {
		return err
	}
	name = fmt.Sprintf("%dx%d_%s.jpg", bounds.Dx(), bounds.Dy(), name)
	filename := filepath.Join(home, "img", name)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, m, nil)
}

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
