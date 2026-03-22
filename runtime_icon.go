package tankbattle

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/icons/icon_final.png
var windowIconPNG []byte

func setWindowIcon() {
	if len(windowIconPNG) == 0 {
		return
	}
	img, err := png.Decode(bytes.NewReader(windowIconPNG))
	if err != nil {
		return
	}
	ebiten.SetWindowIcon(buildWindowIcons(img))
}

func buildWindowIcons(src image.Image) []image.Image {
	sizes := []int{16, 32, 48, 64, 128}
	icons := make([]image.Image, 0, len(sizes))
	for _, size := range sizes {
		icons = append(icons, resizeIcon(src, size))
	}
	return icons
}

func resizeIcon(src image.Image, size int) image.Image {
	srcBounds := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	scaleX := float64(srcBounds.Dx()) / float64(size)
	scaleY := float64(srcBounds.Dy()) / float64(size)

	for y := 0; y < size; y++ {
		srcY0 := int(math.Floor(float64(y) * scaleY))
		srcY1 := int(math.Ceil(float64(y+1) * scaleY))
		if srcY1 <= srcY0 {
			srcY1 = srcY0 + 1
		}
		for x := 0; x < size; x++ {
			srcX0 := int(math.Floor(float64(x) * scaleX))
			srcX1 := int(math.Ceil(float64(x+1) * scaleX))
			if srcX1 <= srcX0 {
				srcX1 = srcX0 + 1
			}
			dst.Set(x, y, averageColor(src, srcBounds.Min.X+srcX0, srcBounds.Min.Y+srcY0, srcBounds.Min.X+srcX1, srcBounds.Min.Y+srcY1))
		}
	}

	return dst
}

func averageColor(src image.Image, x0, y0, x1, y1 int) color.NRGBA {
	var rSum, gSum, bSum, aSum uint64
	var count uint64

	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			r, g, b, a := src.At(x, y).RGBA()
			rSum += uint64(r >> 8)
			gSum += uint64(g >> 8)
			bSum += uint64(b >> 8)
			aSum += uint64(a >> 8)
			count++
		}
	}

	if count == 0 {
		return color.NRGBA{}
	}

	return color.NRGBA{
		R: uint8(rSum / count),
		G: uint8(gSum / count),
		B: uint8(bSum / count),
		A: uint8(aSum / count),
	}
}
