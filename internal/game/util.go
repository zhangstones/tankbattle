package game

import "image/color"

func toRGBA(c color.Color) color.RGBA {
	r, g, b, a := c.RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

func shift(c color.RGBA, dr, dg, db int) color.RGBA {
	r := clampInt(int(c.R)+dr, 0, 255)
	g := clampInt(int(c.G)+dg, 0, 255)
	b := clampInt(int(c.B)+db, 0, 255)
	return color.RGBA{uint8(r), uint8(g), uint8(b), c.A}
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func tankRect(t tank) rect { return rect{t.x, t.y, tankSize, tankSize} }

func overlap(a, b rect) bool {
	return a.x < b.x+b.w && a.x+a.w > b.x && a.y < b.y+b.h && a.y+a.h > b.y
}

func onOffText(on bool) string {
	if on {
		return "ON"
	}
	return "OFF"
}

func (g *game) setMessage(s string, tick int) {
	g.msg = s
	g.msgTick = tick
}
