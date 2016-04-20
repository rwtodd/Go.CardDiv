package main // import "go.waywardcode.com/carddiv"

import (
	"image"
	"image/color"
)

// Here are some rotated image structs (just for 180 and 90 degrees).
// They are implemented by wrapping an image.Image, and translating
// coordinates on the fly.

// A revesedCard is flipped 180 degrees.
type reversedCard struct {
	image.Image
}

func (rc *reversedCard) At(x, y int) color.Color {
	var b = rc.Bounds()
	return rc.Image.At(b.Max.X-x+b.Min.X,
		b.Max.Y-y+b.Min.Y)
}

// A sideways card is flipped 90 degrees counter-clockwise.
type sidewaysCard struct {
	image.Image
}

// The bounds of a sideways card are swapped in X and Y.
func (sc *sidewaysCard) Bounds() image.Rectangle {
	var orig = sc.Image.Bounds()
	return image.Rect(orig.Min.Y, orig.Min.X, orig.Max.Y, orig.Max.X)
}

// Translate the points to the base image turned on its side.
func (sc *sidewaysCard) At(x, y int) color.Color {
	var b = sc.Image.Bounds()
	return sc.Image.At(b.Max.X - y + b.Min.X, x)
}
