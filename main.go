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
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	_ "image/gif"
	_ "image/png"
)

func main() {
	var (
		dir     = flag.String("dir", "", "directory to save file (default: cwd)")
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
	o := &options{
		dir:     *dir,
		name:    *name,
		quality: *quality,
	}
	err := run(args[0], o)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type options struct {
	dir     string
	name    string
	quality int
}

func (o *options) clean(src string) error {
	var err error
	if o.name == "" {
		o.name = baseWithoutExt(src)
	}
	o.name, err = normalize(o.name)
	if err != nil {
		return err
	}
	return nil
}

func run(src string, o *options) error {
	err := o.clean(src)
	if err != nil {
		return err
	}
	resp, err := http.Get(src)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return save(resp.Body, o)
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

func save(r io.Reader, o *options) error {
	m, _, err := image.Decode(r)
	if err != nil {
		return err
	}
	rect := m.Bounds()
	filename := getSaveName(rect, o)
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return jpeg.Encode(f, m, &jpeg.Options{Quality: o.quality})
}

func getSaveName(rect image.Rectangle, o *options) string {
	name := fmt.Sprintf("%dx%d_%s.jpg", rect.Dx(), rect.Dy(), o.name)
	return filepath.Join(o.dir, name)
}
