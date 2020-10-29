package lib

import (
	"fmt"
	"image"
	"image/gif"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imageorient"
)

type FileImage struct {
	Filename string
}

func (f FileImage) Image() (image.Image, error) {
	file, err := os.Open(f.Filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := imageorient.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (f FileImage) Gif() (*gif.GIF, error) {
	ext := filepath.Ext(f.Filename)
	t := mime.TypeByExtension(ext)
	if !strings.Contains(t, "gif") {
		return nil, nil
	}

	file, err := os.Open(f.Filename)
	if err != nil {
		return nil, err
	}

	gimg, err := gif.DecodeAll(file)
	if err != nil {
		return nil, err
	}
	return gimg, nil
}

func (f FileImage) LoadingMsg() string {
	return fmt.Sprintf("Loading %s ✨", f.Filename)
}

func (f FileImage) Footer() string {
	return f.Filename
}
