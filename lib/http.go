package lib

import (
	"fmt"
	"image"
	"image/gif"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/disintegration/imageorient"
)

type HttpImage struct {
	img image.Image
	gif *gif.GIF
	URL string
}

func (h HttpImage) Image() (image.Image, error) {
	if h.img != nil {
		return h.img, nil
	}
	resp, err := http.Get(h.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err := imageorient.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	h.img = img

	return h.img, nil
}

func (h HttpImage) Gif() (*gif.GIF, error) {
	if h.gif != nil {
		return h.gif, nil
	}
	ext := filepath.Ext(h.URL)
	t := mime.TypeByExtension(ext)
	if !strings.Contains(t, "gif") {
		return nil, nil
	}

	resp, err := http.Get(h.URL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	gimg, err := gif.DecodeAll(resp.Body)
	if err != nil {
		return nil, err
	}
	h.gif = gimg
	return h.gif, nil
}

func (h HttpImage) LoadingMsg() string {
	return fmt.Sprintf("Loading %s ✨", h.URL)
}

func (h HttpImage) Footer() string {
	return h.URL
}
