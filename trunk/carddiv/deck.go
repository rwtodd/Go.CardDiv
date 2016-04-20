package main

import (
	"archive/zip"
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"math/rand"
	"strings"
)

// representation for our card deck...

type deck struct {
	name  string
	zfile *zip.ReadCloser
	imgs  []*zip.File
	ratio float64
}

func NewDeck(fn string) (*deck, error) {
	zfile, err := zip.OpenReader(fn)
	if err != nil {
		return nil, err
	}

	// filter the files down to the JPG files...
	imgs := make([]*zip.File, 0, len(zfile.File))
	for _, v := range zfile.File {
		lowName := strings.ToLower(v.FileInfo().Name())
		if v.FileInfo().Mode().IsRegular() &&
			(strings.HasSuffix(lowName, ".jpg") ||
				strings.HasSuffix(lowName, ".jpeg")) {
			imgs = append(imgs, v)
		}
	}

	// now grab the aspect ratio of the files...
	// assuming that one of them is as good as any other...
	if len(imgs) < 1 {
		return nil, fmt.Errorf("No images in the zip file!")
	}
	rat, err := determineRatio(imgs[0])
	if err != nil {
		return nil, err
	}

	// all is well, give back the deck..
	return &deck{fn, zfile, imgs, rat}, nil
}

func determineRatio(zimg *zip.File) (float64, error) {
	img, err := zimg.Open()
	if err != nil {
		return 0.0, err
	}
	defer img.Close()

	dec, err := jpeg.Decode(img)
	if err != nil {
		return 0.0, err
	}
	b := dec.Bounds()
	return (float64(b.Dx()) / float64(b.Dy())), nil
}

func (dk *deck) Name() string { return dk.name }

func (dk *deck) NumCards() int { return len(dk.imgs) }

func (dk *deck) CardHeight(width uint) uint { return uint(float64(width) / dk.ratio) }

func (dk *deck) Close() error { return dk.zfile.Close() }

type cardOpts struct {
	reversed bool
	onSide   bool
}

func (dk *deck) Image(which int, width uint, options cardOpts) (image.Image, error) {
	if which > len(dk.imgs) {
		return nil, fmt.Errorf("%d is greater than %d images in deck!", which, len(dk.imgs))
	}

	img, err := dk.imgs[which].Open()
	if err != nil {
		return nil, err
	}
	defer img.Close()

	var cardImg image.Image
	cardImg, err = jpeg.Decode(img)
	if err != nil {
		return nil, err
	}

	// resize image ...
	cardImg = resize.Resize(width, dk.CardHeight(width), cardImg, resize.Bicubic)

	// possibly rotate the image...
	if options.reversed {
		cardImg = &reversedCard{cardImg}
	}

	if options.onSide {
		cardImg = &sidewaysCard{cardImg}
	}

	return cardImg, nil
}

func (dk *deck) Shuffled(howMany int) ([]int, error) {
	dsize := len(dk.imgs)
	if howMany > dsize {
		return nil, fmt.Errorf("Not enough cards in deck to get %d.", howMany)
	}
	shuffled := rand.Perm(dsize)
	return shuffled[:howMany], nil
}
