package ui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

var (
	uiInk            = color.RGBA{8, 14, 20, 236}
	uiPanelFill      = color.RGBA{10, 21, 31, 220}
	uiPanelSoft      = color.RGBA{18, 34, 48, 188}
	uiPanelLine      = color.RGBA{92, 166, 176, 140}
	uiPanelEdge      = color.RGBA{18, 220, 206, 70}
	uiSteelBlue      = color.RGBA{96, 174, 208, 220}
	uiSignalGreen    = color.RGBA{82, 224, 148, 232}
	uiSignalAmber    = color.RGBA{255, 194, 84, 232}
	uiSignalRed      = color.RGBA{255, 106, 88, 232}
	uiSignalOrange   = color.RGBA{255, 152, 92, 224}
	uiMutedFill      = color.RGBA{18, 24, 32, 218}
	uiMutedLine      = color.RGBA{56, 80, 98, 180}
	uiBackgroundTop  = color.RGBA{8, 18, 30, 255}
	uiBackgroundMid  = color.RGBA{17, 36, 44, 255}
	uiBackgroundBase = color.RGBA{38, 52, 44, 255}
)

func alpha(c color.RGBA, a uint8) color.RGBA {
	c.A = a
	return c
}

func blend(a, b color.RGBA, t float64) color.RGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	lerp := func(x, y uint8) uint8 {
		return uint8(float64(x) + (float64(y)-float64(x))*t)
	}
	return color.RGBA{
		R: lerp(a.R, b.R),
		G: lerp(a.G, b.G),
		B: lerp(a.B, b.B),
		A: lerp(a.A, b.A),
	}
}

func pulse(frame int, speed float64, lo, hi float64) float64 {
	if hi < lo {
		lo, hi = hi, lo
	}
	wave := (math.Sin(float64(frame)*speed) + 1) * 0.5
	return lo + (hi-lo)*wave
}

func drawRectOutline(screen *ebiten.Image, x, y, w, h, thick float64, c color.Color) {
	if thick <= 0 {
		return
	}
	ebitenutil.DrawRect(screen, x, y, w, thick, c)
	ebitenutil.DrawRect(screen, x, y+h-thick, w, thick, c)
	ebitenutil.DrawRect(screen, x, y, thick, h, c)
	ebitenutil.DrawRect(screen, x+w-thick, y, thick, h, c)
}

func drawBracketCorners(screen *ebiten.Image, x, y, w, h, size float64, c color.Color) {
	ebitenutil.DrawRect(screen, x, y, size, 1, c)
	ebitenutil.DrawRect(screen, x, y, 1, size, c)
	ebitenutil.DrawRect(screen, x+w-size, y, size, 1, c)
	ebitenutil.DrawRect(screen, x+w-1, y, 1, size, c)
	ebitenutil.DrawRect(screen, x, y+h-1, size, 1, c)
	ebitenutil.DrawRect(screen, x, y+h-size, 1, size, c)
	ebitenutil.DrawRect(screen, x+w-size, y+h-1, size, 1, c)
	ebitenutil.DrawRect(screen, x+w-1, y+h-size, 1, size, c)
}

func drawSurfacePanel(screen *ebiten.Image, x, y, w, h float64, accent color.RGBA) {
	drawGlow(screen, x-2, y-2, w+4, h+4, 2, alpha(accent, 12))
	ebitenutil.DrawRect(screen, x, y, w, h, uiInk)
	ebitenutil.DrawRect(screen, x+2, y+2, w-4, h-4, uiPanelFill)
	ebitenutil.DrawRect(screen, x+2, y+2, w-4, 14, alpha(uiPanelSoft, 210))
	drawRectOutline(screen, x+1, y+1, w-2, h-2, 1, alpha(uiPanelLine, 160))
	ebitenutil.DrawRect(screen, x+12, y+8, w-24, 1, alpha(accent, 170))
	drawBracketCorners(screen, x+1, y+1, w-2, h-2, 10, alpha(accent, 210))
}

func drawInsetPanel(screen *ebiten.Image, x, y, w, h float64, accent color.RGBA, selected bool, frame int) {
	fill := uiMutedFill
	line := uiMutedLine
	if selected {
		fill = alpha(blend(uiPanelSoft, accent, 0.16), 228)
		line = alpha(accent, uint8(pulse(frame, 0.12, 160, 235)))
		drawGlow(screen, x-1, y-1, w+2, h+2, 1, alpha(accent, 18))
	}
	ebitenutil.DrawRect(screen, x, y, w, h, fill)
	drawRectOutline(screen, x, y, w, h, 1, line)
	ebitenutil.DrawRect(screen, x+1, y+1, w-2, 1, alpha(accent, 75))
	if selected {
		ebitenutil.DrawRect(screen, x+5, y+5, 4, h-10, line)
	}
}

func drawPill(screen *ebiten.Image, x, y, w, h float64, accent color.RGBA) {
	ebitenutil.DrawRect(screen, x, y, w, h, alpha(uiInk, 228))
	drawRectOutline(screen, x, y, w, h, 1, alpha(accent, 210))
	ebitenutil.DrawRect(screen, x+1, y+1, w-2, 1, alpha(accent, 96))
}

func drawMeter(screen *ebiten.Image, x, y, w, h, rate float64, fill color.RGBA) {
	if rate < 0 {
		rate = 0
	}
	if rate > 1 {
		rate = 1
	}
	ebitenutil.DrawRect(screen, x, y, w, h, alpha(uiInk, 232))
	drawRectOutline(screen, x, y, w, h, 1, alpha(uiPanelLine, 150))
	innerW := w - 4
	if innerW < 0 {
		innerW = 0
	}
	fillW := innerW * rate
	ebitenutil.DrawRect(screen, x+2, y+2, fillW, h-4, fill)
	if fillW > 0 {
		ebitenutil.DrawRect(screen, x+2, y+2, fillW, 2, alpha(shift(fill, 26, 26, 26), 160))
	}
}

func drawGlow(screen *ebiten.Image, x, y, w, h float64, steps int, c color.RGBA) {
	if steps < 1 {
		steps = 1
	}
	for i := 0; i < steps; i++ {
		pad := float64(i * 2)
		a := uint8(float64(c.A) * (1 - float64(i)/float64(steps+1)))
		ebitenutil.DrawRect(screen, x-pad, y-pad, w+pad*2, h+pad*2, alpha(c, a))
	}
}
