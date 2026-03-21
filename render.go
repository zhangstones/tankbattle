package tankbattle

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func (g *game) Draw(screen *ebiten.Image) {
	drawBackground(screen)

	if g.state == stateMenu {
		drawMenu(screen, g)
		return
	}

	for _, w := range g.walls {
		drawWall(screen, w)
	}
	drawFortress(screen, g.fort)

	if g.shieldTick > 0 {
		drawCircle(screen, g.player.x+tankSize/2, g.player.y+tankSize/2, 24, color.RGBA{86, 182, 255, 55})
	}
	drawTank(screen, g.player, color.RGBA{48, 206, 132, 255}, color.RGBA{160, 255, 206, 255})
	for _, e := range g.enemies {
		body := color.RGBA{226, 86, 56, 255}
		accent := color.RGBA{255, 190, 132, 255}
		if e.role == roleLeftFlank || e.role == roleRightFlank {
			body = color.RGBA{210, 120, 62, 255}
			accent = color.RGBA{255, 220, 160, 255}
		}
		drawTank(screen, *e, body, accent)
	}

	for _, b := range g.bullets {
		core := color.RGBA{255, 244, 177, 255}
		glow := color.RGBA{255, 170, 64, 120}
		if !b.fromPlayer {
			core = color.RGBA{255, 149, 149, 255}
			glow = color.RGBA{255, 88, 88, 110}
		}
		ebitenutil.DrawRect(screen, b.x-2, b.y-2, bulletSize+4, bulletSize+4, glow)
		ebitenutil.DrawRect(screen, b.x, b.y, bulletSize, bulletSize, core)
	}

	for _, p := range g.powerups {
		drawPowerup(screen, p)
	}
	for _, ex := range g.explosions {
		progress := 1 - float64(ex.life)/float64(ex.max)
		r := ex.radius * (0.35 + progress)
		alpha := uint8(float64(220) * (1 - progress))
		drawCircle(screen, ex.x, ex.y, r, color.RGBA{255, 186, 92, alpha})
		drawCircle(screen, ex.x, ex.y, r*0.55, color.RGBA{255, 236, 160, alpha})
	}

	drawHUD(screen, g)

	if g.msg != "" {
		ebitenutil.DrawRect(screen, screenW/2-160, 84, 320, 34, color.RGBA{8, 14, 18, 220})
		ebitenutil.DrawRect(screen, screenW/2-156, 88, 312, 26, color.RGBA{44, 104, 118, 120})
		ebitenutil.DebugPrintAt(screen, g.msg, screenW/2-58, 96)
	}
	if g.paused {
		ebitenutil.DrawRect(screen, screenW/2-98, screenH/2-24, 196, 48, color.RGBA{10, 15, 20, 220})
		ebitenutil.DrawRect(screen, screenW/2-94, screenH/2-20, 188, 40, color.RGBA{58, 74, 92, 120})
		ebitenutil.DebugPrintAt(screen, "Paused [P] Resume", screenW/2-54, screenH/2-4)
	}
	if g.state == stateEnded {
		msg := "Defeat - Press R to Restart"
		if g.win {
			msg = "Victory - Fortress Survived"
		}
		ebitenutil.DrawRect(screen, screenW/2-220, screenH/2-45, 440, 90, color.RGBA{12, 16, 22, 220})
		ebitenutil.DebugPrintAt(screen, msg, screenW/2-100, screenH/2-12)
		ebitenutil.DebugPrintAt(screen, "R restart  M menu", screenW/2-54, screenH/2+12)
	}
}

func drawMenu(screen *ebiten.Image, g *game) {
	ebitenutil.DrawRect(screen, 130, 80, 700, 480, color.RGBA{8, 14, 20, 235})
	ebitenutil.DrawRect(screen, 136, 86, 688, 468, color.RGBA{30, 64, 74, 130})
	ebitenutil.DrawRect(screen, 164, 116, 632, 62, color.RGBA{16, 92, 90, 120})
	ebitenutil.DebugPrintAt(screen, "TANK BATTLE // MISSION SETTINGS", 320, 140)
	ebitenutil.DebugPrintAt(screen, "UP/DOWN select, LEFT/RIGHT modify, ENTER start", 250, 198)
	ebitenutil.DebugPrintAt(screen, "Shortcuts: 1/2/3 difficulty, +/- enemy amount", 272, 218)
	ebitenutil.DebugPrintAt(screen, "Combat: hold WASD strafe, double-tap WASD turn, fire J/Space", 210, 238)

	diffText := "Normal"
	diffDesc := "Balanced speed and enemy fire rate."
	if g.difficulty == diffEasy {
		diffText = "Easy"
		diffDesc = "Slower enemies, lower pressure."
	} else if g.difficulty == diffHard {
		diffText = "Hard"
		diffDesc = "Faster enemies with higher HP."
	}

	titles := []string{
		"Difficulty: " + diffText,
		fmt.Sprintf("Enemy Count / Wave: %d", g.enemyBase),
		"Start Mission",
	}
	descs := []string{
		diffDesc,
		"How many tanks appear each wave.",
		"Start with current settings.",
	}

	for i := 0; i < len(titles); i++ {
		y := 286 + i*74
		bg := color.RGBA{20, 34, 40, 180}
		if g.menuIndex == i {
			bg = color.RGBA{72, 138, 100, 170}
		}
		ebitenutil.DrawRect(screen, 214, float64(y-14), 532, 58, bg)
		ebitenutil.DebugPrintAt(screen, titles[i], 246, y)
		ebitenutil.DebugPrintAt(screen, descs[i], 246, y+20)
	}

	ebitenutil.DrawRect(screen, 164, 516, 632, 28, color.RGBA{18, 26, 34, 220})
	ebitenutil.DebugPrintAt(screen, "Tip: Press [R] to restart instantly, [M] to return menu.", 232, 524)
}

func drawHUD(screen *ebiten.Image, g *game) {
	line1 := fmt.Sprintf("HP:%d   SCORE:%d   ENEMY:%d   WAVE:%d/%d", g.player.hp, g.score, len(g.enemies), g.wave, g.maxWave)
	line2 := "Hold WASD strafe  Double-tap WASD turn  Fire J/Space"
	line3 := fmt.Sprintf("BUFF  SHIELD:%2ds   RAPID:%2ds", g.shieldTick/60, g.rapidTick/60)

	textW := maxInt(textWidth(line1), maxInt(textWidth(line2), textWidth(line3)))
	panelW := clampInt(textW+56, 420, 620)
	badgeX := panelW - 96

	ebitenutil.DrawRect(screen, 10, 10, float64(panelW), 96, color.RGBA{8, 16, 22, 220})
	ebitenutil.DrawRect(screen, 14, 14, float64(panelW-8), 88, color.RGBA{40, 86, 96, 135})
	ebitenutil.DebugPrintAt(screen, line1, 24, 22)
	ebitenutil.DebugPrintAt(screen, line2, 24, 44)
	ebitenutil.DebugPrintAt(screen, line3, 24, 66)

	if g.shieldTick > 0 {
		ebitenutil.DrawRect(screen, float64(badgeX), 66, 82, 20, color.RGBA{66, 120, 200, 190})
		ebitenutil.DebugPrintAt(screen, "SHIELD", badgeX+10, 72)
	} else if g.rapidTick > 0 {
		ebitenutil.DrawRect(screen, float64(badgeX), 66, 82, 20, color.RGBA{200, 146, 56, 190})
		ebitenutil.DebugPrintAt(screen, "RAPID", badgeX+12, 72)
	}

	barX, barY, barW, barH := 700.0, 30.0, 180.0, 12.0
	ebitenutil.DebugPrintAt(screen, "FORTRESS", int(barX), int(barY)-16)
	ebitenutil.DrawRect(screen, barX, barY, barW, barH, color.RGBA{40, 40, 46, 220})
	rate := float64(g.fort.hp) / float64(g.fort.maxHP)
	if rate < 0 {
		rate = 0
	}
	fill := color.RGBA{69, 220, 148, 240}
	if rate < 0.45 {
		fill = color.RGBA{240, 96, 74, 240}
	}
	ebitenutil.DrawRect(screen, barX+1, barY+1, (barW-2)*rate, barH-2, fill)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d/%d", g.fort.hp, g.fort.maxHP), int(barX)+int(barW)+8, int(barY)-1)
}

func drawBackground(screen *ebiten.Image) {
	for y := 0; y < screenH; y += 3 {
		t := float64(y) / float64(screenH)
		r := uint8(34 + 22*t)
		g := uint8(44 + 18*t)
		b := uint8(46 + 8*t)
		ebitenutil.DrawRect(screen, 0, float64(y), screenW, 3, color.RGBA{r, g, b, 255})
	}
	for y := 28; y < screenH; y += 30 {
		ebitenutil.DrawLine(screen, 0, float64(y), screenW, float64(y), color.RGBA{72, 64, 54, 35})
	}
	for x := 20; x < screenW; x += 120 {
		for y := 0; y < screenH; y += 48 {
			ebitenutil.DrawRect(screen, float64(x), float64(y), 36, 4, color.RGBA{58, 54, 48, 45})
			ebitenutil.DrawRect(screen, float64(x+46), float64(y+14), 36, 4, color.RGBA{58, 54, 48, 36})
		}
	}
}

func drawWall(screen *ebiten.Image, w *wall) {
	if w.guard {
		drawGuardWall(screen, w)
		return
	}

	if w.destructive {
		rate := float64(w.hp) / float64(w.maxHP)
		if rate < 0 {
			rate = 0
		}
		base := color.RGBA{150, 92, 64, 255}
		dim := int((1 - rate) * 35)
		base = shift(base, -dim, -dim, -dim)
		ebitenutil.DrawRect(screen, w.box.x, w.box.y, w.box.w, w.box.h, base)
		if w.box.w <= tankSize+0.1 || w.box.h <= tankSize+0.1 {
			// Chunk walls keep a clean edge only; skip inner brick seams to avoid stray vertical lines.
			ebitenutil.DrawRect(screen, w.box.x+1, w.box.y+1, maxF(w.box.w-2, 1), 1, color.RGBA{206, 166, 122, 70})
			ebitenutil.DrawRect(screen, w.box.x+1, w.box.y+w.box.h-2, maxF(w.box.w-2, 1), 1, color.RGBA{78, 58, 44, 90})
			return
		}
		brickH := 8.0
		brickW := 16.0
		for y := w.box.y; y < w.box.y+w.box.h; y += brickH {
			ebitenutil.DrawLine(screen, w.box.x, y, w.box.x+w.box.w, y, color.RGBA{96, 84, 72, 135})
			row := int((y - w.box.y) / brickH)
			offset := 0.0
			if row%2 == 1 {
				offset = brickW / 2
			}
			for x := w.box.x - offset; x < w.box.x+w.box.w; x += brickW {
				if x > w.box.x && x < w.box.x+w.box.w {
					ebitenutil.DrawLine(screen, x, y, x, y+brickH, color.RGBA{98, 86, 74, 125})
				}
			}
		}
		ebitenutil.DrawRect(screen, w.box.x+1, w.box.y+1, w.box.w-2, 2, color.RGBA{206, 166, 122, 70})
		if rate < 0.75 {
			ebitenutil.DrawLine(screen, w.box.x+4, w.box.y+4, w.box.x+w.box.w*0.55, w.box.y+w.box.h-3, color.RGBA{60, 42, 34, 180})
		}
		if rate < 0.45 {
			ebitenutil.DrawLine(screen, w.box.x+w.box.w-6, w.box.y+4, w.box.x+w.box.w*0.35, w.box.y+w.box.h-2, color.RGBA{56, 38, 28, 200})
		}
		return
	}

	base := color.RGBA{108, 114, 120, 255}
	ebitenutil.DrawRect(screen, w.box.x, w.box.y, w.box.w, w.box.h, base)
	ebitenutil.DrawRect(screen, w.box.x+1, w.box.y+1, w.box.w-2, 2, color.RGBA{190, 198, 208, 64})
	ebitenutil.DrawRect(screen, w.box.x+1, w.box.y+w.box.h-3, w.box.w-2, 2, color.RGBA{52, 56, 62, 90})
	for x := w.box.x + 10; x < w.box.x+w.box.w-6; x += 18 {
		drawCircle(screen, x, w.box.y+w.box.h/2, 1.8, color.RGBA{188, 192, 200, 130})
	}
}

func drawGuardWall(screen *ebiten.Image, w *wall) {
	base := color.RGBA{128, 100, 74, 255}
	ebitenutil.DrawRect(screen, w.box.x, w.box.y, w.box.w, w.box.h, base)
	ebitenutil.DrawRect(screen, w.box.x+1, w.box.y+1, maxF(w.box.w-2, 1), 1, color.RGBA{194, 158, 126, 70})
	ebitenutil.DrawRect(screen, w.box.x+1, w.box.y+w.box.h-2, maxF(w.box.w-2, 1), 1, color.RGBA{74, 56, 42, 90})
}

func drawFortress(screen *ebiten.Image, fort fortress) {
	cx := fort.box.x + fort.box.w/2
	drawCircle(screen, cx, fort.box.y+fort.box.h/2, 48, color.RGBA{140, 166, 185, 24})
	ebitenutil.DrawRect(screen, fort.box.x, fort.box.y, fort.box.w, fort.box.h, color.RGBA{72, 78, 86, 255})
	ebitenutil.DrawRect(screen, fort.box.x+3, fort.box.y+3, fort.box.w-6, fort.box.h-6, color.RGBA{114, 124, 136, 255})
	ebitenutil.DrawRect(screen, fort.box.x+8, fort.box.y+fort.box.h/2-3, fort.box.w-16, 6, color.RGBA{42, 48, 54, 220})
	drawCircle(screen, cx, fort.box.y+fort.box.h/2, 5, color.RGBA{170, 184, 196, 180})
}

func drawPowerup(screen *ebiten.Image, p *powerup) {
	base := color.RGBA{112, 190, 255, 230}
	label := "S"
	switch p.kind {
	case powerRapid:
		base = color.RGBA{255, 194, 84, 230}
		label = "R"
	case powerRepair:
		base = color.RGBA{118, 225, 146, 230}
		label = "F"
	}
	ebitenutil.DrawRect(screen, p.box.x-2, p.box.y-2, p.box.w+4, p.box.h+4, color.RGBA{22, 28, 34, 180})
	ebitenutil.DrawRect(screen, p.box.x, p.box.y, p.box.w, p.box.h, base)
	ebitenutil.DebugPrintAt(screen, label, int(p.box.x+4), int(p.box.y+3))
}

func (g *game) Layout(_, _ int) (int, int) { return screenW, screenH }

func drawTank(screen *ebiten.Image, t tank, body, accent color.Color) {
	cx := t.x + tankSize/2
	cy := t.y + tankSize/2
	trackDark := color.RGBA{34, 38, 42, 255}
	trackLight := color.RGBA{58, 62, 68, 255}
	hullShadow := shift(toRGBA(body), -30, -30, -30)
	turret := shift(toRGBA(accent), -8, -8, -8)
	barrel := color.RGBA{210, 214, 220, 255}

	ebitenutil.DrawRect(screen, t.x+1, t.y+2, 8, tankSize-4, trackDark)
	ebitenutil.DrawRect(screen, t.x+tankSize-9, t.y+2, 8, tankSize-4, trackDark)
	for i := 0; i < 4; i++ {
		wy := t.y + 6 + float64(i)*7
		drawCircle(screen, t.x+5, wy, 1.7, trackLight)
		drawCircle(screen, t.x+tankSize-5, wy, 1.7, trackLight)
	}
	ebitenutil.DrawRect(screen, t.x+7, t.y+5, tankSize-14, tankSize-10, hullShadow)
	ebitenutil.DrawRect(screen, t.x+8, t.y+6, tankSize-16, tankSize-12, body)
	ebitenutil.DrawRect(screen, t.x+10, t.y+8, tankSize-20, 4, shift(toRGBA(body), 20, 20, 20))

	drawCircle(screen, cx, cy, 8.5, turret)
	drawCircle(screen, cx, cy, 6.2, accent)
	drawCircle(screen, cx, cy, 2.2, color.RGBA{68, 72, 80, 255})

	switch t.turret {
	case up:
		ebitenutil.DrawRect(screen, cx-2.5, t.y-11, 5, 16, barrel)
		ebitenutil.DrawRect(screen, cx-4, t.y-13, 8, 3, color.RGBA{138, 144, 152, 255})
	case down:
		ebitenutil.DrawRect(screen, cx-2.5, t.y+tankSize-5, 5, 16, barrel)
		ebitenutil.DrawRect(screen, cx-4, t.y+tankSize+10, 8, 3, color.RGBA{138, 144, 152, 255})
	case left:
		ebitenutil.DrawRect(screen, t.x-11, cy-2.5, 16, 5, barrel)
		ebitenutil.DrawRect(screen, t.x-13, cy-4, 3, 8, color.RGBA{138, 144, 152, 255})
	case right:
		ebitenutil.DrawRect(screen, t.x+tankSize-5, cy-2.5, 16, 5, barrel)
		ebitenutil.DrawRect(screen, t.x+tankSize+10, cy-4, 3, 8, color.RGBA{138, 144, 152, 255})
	}
}

func drawCircle(screen *ebiten.Image, cx, cy, r float64, c color.Color) {
	r2 := r * r
	for y := -int(r); y <= int(r); y++ {
		fy := float64(y)
		radicand := r2 - fy*fy
		if radicand <= 0 {
			continue
		}
		dx := math.Sqrt(radicand)
		ebitenutil.DrawRect(screen, cx-dx, cy+fy, dx*2, 1, c)
	}
}

func textWidth(s string) int {
	return len([]rune(s)) * 7
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
