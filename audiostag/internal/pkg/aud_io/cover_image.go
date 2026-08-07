package aud_io

import "image"

type CoverImage struct {
	Filename string
	image.Image
}

func (c *CoverImage) Set(filename string, img image.Image) {
	c.Filename = filename
	c.Image = img

}
