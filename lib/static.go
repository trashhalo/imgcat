package lib

import (
	"image"
)

type staticImage struct {
	img image.Image
}

func (s staticImage) Image() (image.Image, error) {
	return s.img, nil
}
