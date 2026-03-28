package testkit

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

func DiffPNG(goldenPath, actualPath string) (bool, image.Image, error) {
	golden, err := decodePNG(goldenPath)
	if err != nil {
		return false, nil, fmt.Errorf("decode golden: %w", err)
	}
	actual, err := decodePNG(actualPath)
	if err != nil {
		return false, nil, fmt.Errorf("decode actual: %w", err)
	}
	if !golden.Bounds().Eq(actual.Bounds()) {
		return false, diffSizeMismatch(golden.Bounds(), actual.Bounds()), nil
	}
	bounds := golden.Bounds()
	diff := image.NewRGBA(bounds)
	match := true
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			g := color.RGBAModel.Convert(golden.At(x, y)).(color.RGBA)
			a := color.RGBAModel.Convert(actual.At(x, y)).(color.RGBA)
			if g == a {
				diff.SetRGBA(x, y, color.RGBA{R: 0, G: 0, B: 0, A: 255})
				continue
			}
			match = false
			diff.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 255, A: 255})
		}
	}
	return match, diff, nil
}

func decodePNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

func diffSizeMismatch(golden, actual image.Rectangle) image.Image {
	maxW := max(golden.Dx(), actual.Dx())
	maxH := max(golden.Dy(), actual.Dy())
	diff := image.NewRGBA(image.Rect(0, 0, maxW, maxH))
	fill(diff, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	return diff
}

func fill(img *image.RGBA, c color.RGBA) {
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i+0] = c.R
		img.Pix[i+1] = c.G
		img.Pix[i+2] = c.B
		img.Pix[i+3] = c.A
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
