package ui

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	hudTopY       = 10
	hudHeight     = 118
	hudFrameInset = 2
	statusInset   = 2
	hudMessageGap = 14

	pausePanelW = 340
	pausePanelH = 96
	endPanelW   = 430
	endPanelH   = 124

	menuPanelX = 48
	menuPanelY = 28
	menuPanelW = 864
	menuPanelH = 604

	menuInnerInset = 6
	menuHeaderX    = 92
	menuHeaderY    = 60
	menuHeaderW    = 512
	menuHeaderH    = 76

	menuFooterY = 582

	menuHelpLineGap          = 18
	menuHelpLineCount        = 3
	menuHelpTopGapFromHeader = 18
	menuHelpBottomPadding    = 12
	menuHelpToOptionGap      = 12

	menuOptionBoxHeight     = 56
	menuOptionTextTopInset  = 8
	menuOptionMinGap        = 8
	menuOptionMaxGap        = 18
	menuOptionBottomPadding = 12

	menuTextHeight = 16

	historyPanelX = 72
	historyPanelY = 52
	historyPanelW = 816
	historyPanelH = 556
)

const (
	menuTitleText = "TANK BATTLE // MISSION SETTINGS"
	menuHelpLine1 = "UP/DOWN select, LEFT/RIGHT modify/toggle, ENTER start, FIRE J/Space"
	menuHelpLine2 = "Shortcuts: 1/2/3 difficulty"
	menuHelpLine3 = "Combat: hold WASD/Arrow strafe, double-tap WASD/Arrow turn, FIRE J/Space"
)

func (g *game) Draw(screen *ebiten.Image) {
	drawBackground(screen, g)

	if g.state == stateMenu {
		drawMenu(screen, g)
		return
	}

	for _, w := range g.walls {
		drawWall(screen, w)
	}
	drawFortress(screen, g.fort)

	if g.shieldTick > 0 {
		pulseAlpha := uint8(pulse(g.audioFrame, 0.18, 48, 86))
		drawCircle(screen, g.player.x+tankSize/2, g.player.y+tankSize/2, 24, color.RGBA{86, 182, 255, pulseAlpha})
		drawCircle(screen, g.player.x+tankSize/2, g.player.y+tankSize/2, 29, color.RGBA{86, 182, 255, pulseAlpha / 2})
	}
	drawTank(screen, g.player, color.RGBA{52, 212, 148, 255}, color.RGBA{180, 255, 218, 255})
	for _, e := range g.enemies {
		body := color.RGBA{228, 88, 58, 255}
		accent := color.RGBA{255, 196, 136, 255}
		if e.role == roleLeftFlank || e.role == roleRightFlank {
			body = color.RGBA{214, 126, 68, 255}
			accent = color.RGBA{255, 224, 162, 255}
		}
		drawTank(screen, *e, body, accent)
	}

	for _, b := range g.bullets {
		drawBullet(screen, b)
	}

	for _, p := range g.powerups {
		drawPowerup(screen, p)
	}
	for _, ex := range g.explosions {
		drawExplosion(screen, ex, g.audioFrame)
	}

	drawHUD(screen, g)

	if g.showHistory {
		drawHistoryPanel(screen, g)
		return
	}

	if g.msg != "" {
		msgY := float64(messageBoxTopY())
		drawInsetPanel(screen, float64(screenW/2-190), msgY, 380, 38, uiSteelBlue, true, g.audioFrame)
		ebitenutil.DebugPrintAt(screen, g.msg, centeredTextX(g.msg, screenW/2-190, 380), int(msgY)+12)
	}
	if g.paused {
		drawStatusPanel(screen, pausePanelW, pausePanelH, uiSteelBlue, "Paused", "P resume  M menu")
	}
	if g.state == stateEnded {
		if g.win {
			drawStatusPanel(screen, endPanelW, endPanelH, uiSignalGreen, "Victory", "Fortress survived the assault", "R restart  M menu")
		} else {
			drawStatusPanel(screen, endPanelW, endPanelH, uiSignalRed, "Defeat", "Fortress lost or player destroyed", "R restart  M menu")
		}
	}
}

func drawMenu(screen *ebiten.Image, g *game) {
	layout := computeMenuLayout(menuItemCount)

	const (
		menuOptionsX       = 92
		menuOptionsW       = 520
		menuSidebarX       = 640
		menuSidebarW       = 232
		menuValuePillW     = 128
		menuValuePillH     = 22
		menuValueMeterW    = 128
		menuValueColumnGap = 16
	)
	menuValuePillX := float64(menuOptionsX + menuOptionsW - menuValuePillW - menuValueColumnGap)
	menuValueMeterX := float64(menuOptionsX + menuOptionsW - menuValueMeterW - menuValueColumnGap)

	drawSurfacePanel(screen, menuPanelX, menuPanelY, menuPanelW, menuPanelH, uiSteelBlue)
	drawSurfacePanel(screen, menuHeaderX, menuHeaderY, menuHeaderW, menuHeaderH, uiSignalGreen)

	ebitenutil.DebugPrintAt(screen, menuTitleText, menuTitleX(), menuTitleY())
	ebitenutil.DebugPrintAt(screen, "Tune the mission before deployment. Gameplay rules stay unchanged.", 112, 106)

	diffLabel, diffDesc, diffRate := difficultyPresentation(g.difficulty)

	optionTitles := []string{
		"Difficulty",
		"Total Waves",
		"Sound Effects",
		"SFX Volume",
		"Start Mission",
	}
	optionValues := []string{
		diffLabel,
		fmt.Sprintf("%d waves", g.totalWaves),
		onOffText(g.soundEnabled),
		fmt.Sprintf("%d%%", g.soundVolume),
		"Press Enter",
	}
	optionDescs := []string{
		diffDesc,
		"More waves extend the mission and pressure curve.",
		"Toggle all in-game sound cues.",
		"Change sound level in 25% steps.",
		"Launch battle with the current setup.",
	}

	for i := 0; i < len(optionTitles); i++ {
		top := float64(layout.optionBoxTopY[i])
		selected := g.menuIndex == i
		accent := uiSteelBlue
		if i == 4 {
			accent = uiSignalGreen
		}
		drawInsetPanel(screen, menuOptionsX, top, menuOptionsW, menuOptionBoxHeight, accent, selected, g.audioFrame)
		ebitenutil.DebugPrintAt(screen, optionTitles[i], 116, layout.optionTextY[i])
		ebitenutil.DebugPrintAt(screen, optionDescs[i], 116, layout.optionTextY[i]+18)

		drawPillLabel(screen, menuValuePillX, top+10, menuValuePillW, menuValuePillH, accent, optionValues[i])

		switch i {
		case 0:
			drawMeter(screen, menuValueMeterX, top+38, menuValueMeterW, 8, diffRate, uiSignalGreen)
		case 1:
			drawMeter(screen, menuValueMeterX, top+38, menuValueMeterW, 8, float64(g.totalWaves-matchWaveMin)/float64(matchWaveMax-matchWaveMin), uiSteelBlue)
		case 2:
			stateRate := 0.0
			fill := uiSignalRed
			if g.soundEnabled {
				stateRate = 1
				fill = uiSignalGreen
			}
			drawMeter(screen, menuValueMeterX, top+38, menuValueMeterW, 8, stateRate, fill)
		case 3:
			drawMeter(screen, menuValueMeterX, top+38, menuValueMeterW, 8, float64(g.soundVolume)/100, uiSignalAmber)
		case 4:
			drawMeter(screen, menuValueMeterX, top+38, menuValueMeterW, 8, 1, uiSignalGreen)
		}
	}

	drawInsetPanel(screen, menuSidebarX, 112, menuSidebarW, 92, uiSignalGreen, false, g.audioFrame)
	ebitenutil.DebugPrintAt(screen, "PROFILE", menuSidebarX+18, 128)
	ebitenutil.DebugPrintAt(screen, "Mode: "+diffLabel, menuSidebarX+18, 152)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Waves: %d", g.totalWaves), menuSidebarX+18, 170)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Audio: %s %d%%", onOffText(g.soundEnabled), g.soundVolume), menuSidebarX+18, 188)

	drawInsetPanel(screen, menuSidebarX, 218, menuSidebarW, 116, uiSteelBlue, false, g.audioFrame)
	ebitenutil.DebugPrintAt(screen, "CONTROLS", menuSidebarX+18, 234)
	ebitenutil.DebugPrintAt(screen, "Up/Down: browse", menuSidebarX+18, 258)
	ebitenutil.DebugPrintAt(screen, "Left/Right: adjust", menuSidebarX+18, 276)
	ebitenutil.DebugPrintAt(screen, "Enter: start", menuSidebarX+18, 294)
	ebitenutil.DebugPrintAt(screen, "M: back or resume", menuSidebarX+18, 312)

	drawInsetPanel(screen, menuSidebarX, 348, menuSidebarW, 116, uiSignalAmber, false, g.audioFrame)
	ebitenutil.DebugPrintAt(screen, "TIPS", menuSidebarX+18, 364)
	ebitenutil.DebugPrintAt(screen, "1/2/3 set difficulty.", menuSidebarX+18, 388)
	ebitenutil.DebugPrintAt(screen, "R restarts in battle.", menuSidebarX+18, 406)
	ebitenutil.DebugPrintAt(screen, "H opens score history.", menuSidebarX+18, 424)
	if g.menuResumeAvailable {
		if g.menuRequireRestart {
			ebitenutil.DebugPrintAt(screen, "Setup changes restart run.", menuSidebarX+18, 442)
		} else {
			ebitenutil.DebugPrintAt(screen, "Audio changes keep run.", menuSidebarX+18, 442)
		}
	}
}

func difficultyPresentation(d difficulty) (string, string, float64) {
	switch d {
	case diffEasy:
		return "EASY", "Slower enemies and lighter mission pressure.", 0.3
	case diffHard:
		return "HARD", "Faster waves, higher HP, tighter pressure.", 1
	default:
		return "NORMAL", "Balanced pace and enemy fire cadence.", 0.62
	}
}

func Draw(screen *ebiten.Image, snapshot Snapshot) {
	newCompatGame(snapshot).Draw(screen)
}

func Layout(_, _ int) (int, int) { return screenW, screenH }

func drawHUD(screen *ebiten.Image, g *game) {
	drawHUDCompetitive(screen, g)
}

func drawHUDCompetitive(screen *ebiten.Image, g *game) {
	leftX, leftY, leftW, leftH := 12.0, float64(hudTopY), 556.0, 102.0
	rightX, rightY, rightW, rightH := 582.0, float64(hudTopY), 366.0, 102.0

	drawSurfacePanel(screen, leftX, leftY, leftW, leftH, uiSteelBlue)
	drawSurfacePanel(screen, rightX, rightY, rightW, rightH, uiSignalGreen)

	ebitenutil.DebugPrintAt(screen, "MISSION CONTROL", int(leftX)+18, int(leftY)+16)
	diffLabel, _, diffRate := difficultyPresentation(g.difficulty)
	drawPillLabel(screen, leftX+376, leftY+12, 138, 22, uiSteelBlue, "MODE "+diffLabel)

	waveRate := 1.0
	if g.maxWave > 0 {
		waveRate = float64(g.wave) / float64(g.maxWave)
	}
	drawMeter(screen, leftX+18, leftY+34, 214, 12, waveRate, uiSteelBlue)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("WAVE %d/%d", g.wave, maxInt(g.maxWave, 1)), int(leftX)+244, int(leftY)+32)

	cardY := leftY + 58
	cardW := 122.0
	cards := []struct {
		x      float64
		title  string
		value  string
		accent color.RGBA
	}{
		{leftX + 18, "SCORE", fmt.Sprintf("%d", g.score), uiSignalGreen},
		{leftX + 148, "ENEMIES", fmt.Sprintf("%d", len(g.enemies)), uiSignalRed},
		{leftX + 278, "BEST", fmt.Sprintf("%d", g.bestScore()), uiSignalAmber},
		{leftX + 408, "RANK", fmt.Sprintf("#%d", g.currentRank()), uiSteelBlue},
	}
	for _, card := range cards {
		drawInsetPanel(screen, card.x, cardY, cardW, 30, card.accent, false, g.audioFrame)
		ebitenutil.DebugPrintAt(screen, card.title, int(card.x)+10, int(cardY)+7)
		ebitenutil.DebugPrintAt(screen, card.value, int(card.x)+68, int(cardY)+7)
	}

	statusText := "BUFF OFF"
	statusAccent := uiMutedLine
	if g.shieldTick > 0 {
		statusText = fmt.Sprintf("SHIELD %ds", g.shieldTick/60)
		statusAccent = uiSteelBlue
	} else if g.rapidTick > 0 {
		statusText = fmt.Sprintf("RAPID %ds", g.rapidTick/60)
		statusAccent = uiSignalAmber
	}
	drawPillLabel(screen, leftX+376, leftY+34, 138, 18, statusAccent, statusText)

	drawHUDVitals(screen, g, rightX, rightY, rightW, diffRate)
}

func drawHUDVitals(screen *ebiten.Image, g *game, x, y, w, diffRate float64) {
	const (
		vitalBarW      = 178.0
		vitalValueX    = 210
		threatBoxW     = 82.0
		threatBoxH     = 28.0
		threatInsetX   = 18.0
		threatMeterY   = 46.0
		threatHistoryY = 64
	)
	threatX := x + w - threatBoxW - threatInsetX

	ebitenutil.DebugPrintAt(screen, "FORTRESS INTEGRITY", int(x)+18, int(y)+16)
	fortRate := float64(g.fort.hp) / float64(maxInt(g.fort.maxHP, 1))
	if fortRate < 0 {
		fortRate = 0
	}
	fortFill := uiSignalGreen
	if fortRate < 0.45 {
		fortFill = uiSignalRed
	}
	drawEnergyBar(screen, x+18, y+34, vitalBarW, 14, fortRate, fortFill, "", fortRate < 0.3)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d/%d", g.fort.hp, g.fort.maxHP), int(x)+vitalValueX, int(y)+33)

	tankNow, tankMax := playerCombinedEnergy(g.player)
	tankRate := float64(tankNow) / float64(maxInt(tankMax, 1))
	ebitenutil.DebugPrintAt(screen, "PLAYER ARMOR", int(x)+18, int(y)+58)
	drawEnergyBar(screen, x+18, y+76, vitalBarW, 14, tankRate, uiSignalGreen, "", tankRate < 0.28)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d/%d", tankNow, tankMax), int(x)+vitalValueX, int(y)+75)

	drawInsetPanel(screen, threatX, y+14, threatBoxW, threatBoxH, uiSignalAmber, false, g.audioFrame)
	ebitenutil.DebugPrintAt(screen, "THREAT", centeredTextX("THREAT", int(threatX), int(threatBoxW)), centeredTextY(int(y+14), int(threatBoxH), menuTextHeight))
	drawMeter(screen, threatX, y+threatMeterY, threatBoxW, 10, diffRate, uiSignalAmber)
	ebitenutil.DebugPrintAt(screen, "H history", centeredTextX("H history", int(threatX), int(threatBoxW)), int(y)+threatHistoryY)
}

func drawBoldText(screen *ebiten.Image, s string, x, y int) {
	ebitenutil.DebugPrintAt(screen, s, x+1, y)
	ebitenutil.DebugPrintAt(screen, s, x, y)
}

func drawHistoryPanel(screen *ebiten.Image, g *game) {
	const (
		historyPadX        = 28
		historyTitleYOff   = 24
		historyHeaderYOff  = 84
		historyRowsYOff    = 118
		historyFooterYOff  = 30
		historyColRankX    = 0
		historyColScoreX   = 92
		historyColDurX     = 230
		historyColTimeX    = 360
		historyRowBgHeight = 24
	)

	drawSurfacePanel(screen, historyPanelX, historyPanelY, historyPanelW, historyPanelH, uiSteelBlue)
	drawInsetPanel(screen, historyPanelX+18, historyPanelY+18, historyPanelW-36, 42, uiSteelBlue, false, g.audioFrame)
	drawInsetPanel(screen, historyPanelX+18, historyPanelY+70, historyPanelW-36, 26, uiSignalAmber, false, g.audioFrame)
	ebitenutil.DebugPrintAt(screen, "SCORE HISTORY", historyPanelX+historyPadX, historyPanelY+historyTitleYOff)
	ebitenutil.DebugPrintAt(screen, "H hide  Wheel/PgUp/PgDn scroll", historyPanelX+historyPadX+160, historyPanelY+historyTitleYOff)
	ebitenutil.DebugPrintAt(screen, "RANK", historyPanelX+historyPadX+historyColRankX, historyPanelY+historyHeaderYOff)
	ebitenutil.DebugPrintAt(screen, "SCORE", historyPanelX+historyPadX+historyColScoreX, historyPanelY+historyHeaderYOff)
	ebitenutil.DebugPrintAt(screen, "DURATION", historyPanelX+historyPadX+historyColDurX, historyPanelY+historyHeaderYOff)
	ebitenutil.DebugPrintAt(screen, "TIME (LOCAL)", historyPanelX+historyPadX+historyColTimeX, historyPanelY+historyHeaderYOff)

	entries, start := g.visibleRankEntries()
	if len(entries) == 0 {
		drawInsetPanel(screen, historyPanelX+18, historyPanelY+historyRowsYOff-10, historyPanelW-36, 48, uiSteelBlue, false, g.audioFrame)
		ebitenutil.DebugPrintAt(screen, "No historical records yet.", historyPanelX+historyPadX, historyPanelY+historyRowsYOff+8)
		return
	}

	y := historyPanelY + historyRowsYOff
	for i, e := range entries {
		rank := start + i + 1
		rowY := y + i*hudRankingLineGap
		selected := rank == g.currentRank()
		accent := uiSteelBlue
		if selected {
			accent = uiSignalAmber
		}
		drawInsetPanel(screen, historyPanelX+historyPadX-8, float64(rowY-4), historyPanelW-historyPadX*2+16, historyRowBgHeight, accent, selected, g.audioFrame)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("#%02d", rank), historyPanelX+historyPadX+historyColRankX, rowY)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%6d", e.Score), historyPanelX+historyPadX+historyColScoreX, rowY)
		ebitenutil.DebugPrintAt(screen, formatDuration(e.DurationSec), historyPanelX+historyPadX+historyColDurX, rowY)
		ebitenutil.DebugPrintAt(screen, formatScoreTime(e.At), historyPanelX+historyPadX+historyColTimeX, rowY)
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Showing %d-%d / %d", start+1, start+len(entries), len(g.scoreHistory)), historyPanelX+historyPadX, historyPanelY+historyPanelH-historyFooterYOff)
}

func formatScoreTime(ts string) string {
	if ts == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		if len(ts) >= 16 {
			return ts[:16]
		}
		return ts
	}
	return t.Local().Format("2006-01-02 15:04:05")
}

func formatDuration(sec int) string {
	if sec <= 0 {
		return "00:00"
	}
	m := sec / 60
	s := sec % 60
	return fmt.Sprintf("%02d:%02d", m, s)
}

func drawEnergyBar(screen *ebiten.Image, x, y, w, h, rate float64, fill color.Color, label string, alert bool) {
	if label != "" {
		ebitenutil.DebugPrintAt(screen, label, int(x), int(y)-12)
	}
	fillRGBA := toRGBA(fill)
	if alert {
		drawGlow(screen, x-2, y-2, w+4, h+4, 2, alpha(uiSignalRed, 28))
		fillRGBA = uiSignalRed
	}
	drawMeter(screen, x, y, w, h, rate, fillRGBA)
}

func playerCombinedEnergy(p tank) (int, int) {
	now := maxInt(p.hp, 0) + maxInt(p.turretHP, 0)
	max := maxInt(p.maxHP, 1) + maxInt(p.turretMaxHP, 1)
	return now, max
}

func drawBackground(screen *ebiten.Image, g *game) {
	_ = g

	ebitenutil.DrawRect(screen, 0, 0, screenW, screenH, uiBackgroundBase)
	for y := 0; y < screenH; y += 4 {
		t := float64(y) / float64(screenH)
		col := blend(uiBackgroundTop, uiBackgroundBase, t)
		if t > 0.62 {
			col = blend(uiBackgroundBase, uiBackgroundMid, (t-0.62)/0.38)
		}
		ebitenutil.DrawRect(screen, 0, float64(y), screenW, 4, col)
	}

	for y := 0; y <= screenH; y += gridSize {
		ebitenutil.DrawLine(screen, 0, float64(y), screenW, float64(y), alpha(uiGroundGrid, 20))
	}
	for x := 0; x <= screenW; x += gridSize {
		ebitenutil.DrawLine(screen, float64(x), 0, float64(x), screenH, alpha(uiGroundAsh, 12))
	}

	drawGroundScatter(screen)
}

func drawGroundScatter(screen *ebiten.Image) {
	stones := []struct {
		x float64
		y float64
		r float64
		c color.RGBA
	}{
		{54, 72, 2.2, alpha(uiGroundDust, 64)},
		{92, 98, 1.4, alpha(uiGroundStone, 76)},
		{126, 146, 1.8, alpha(uiGroundStone, 72)},
		{188, 96, 1.5, alpha(uiGroundAsh, 70)},
		{262, 138, 2.0, alpha(uiGroundDust, 62)},
		{352, 84, 1.6, alpha(uiGroundStone, 70)},
		{446, 126, 1.7, alpha(uiGroundAsh, 68)},
		{548, 92, 2.1, alpha(uiGroundDust, 60)},
		{632, 144, 1.6, alpha(uiGroundStone, 74)},
		{744, 108, 1.8, alpha(uiGroundDust, 58)},
		{824, 82, 2.0, alpha(uiGroundStone, 66)},
		{884, 136, 1.5, alpha(uiGroundAsh, 70)},
		{118, 232, 1.6, alpha(uiGroundDust, 58)},
		{236, 274, 2.0, alpha(uiGroundStone, 68)},
		{318, 226, 1.4, alpha(uiGroundAsh, 66)},
		{418, 288, 1.9, alpha(uiGroundDust, 60)},
		{526, 238, 1.5, alpha(uiGroundStone, 72)},
		{646, 284, 1.8, alpha(uiGroundAsh, 68)},
		{758, 248, 2.0, alpha(uiGroundDust, 60)},
		{858, 292, 1.7, alpha(uiGroundStone, 70)},
		{104, 398, 1.6, alpha(uiGroundDust, 56)},
		{214, 454, 2.1, alpha(uiGroundStone, 72)},
		{332, 386, 1.5, alpha(uiGroundAsh, 66)},
		{438, 444, 1.9, alpha(uiGroundDust, 58)},
		{548, 402, 1.5, alpha(uiGroundStone, 70)},
		{666, 458, 2.2, alpha(uiGroundDust, 62)},
		{774, 392, 1.7, alpha(uiGroundAsh, 68)},
		{864, 448, 2.0, alpha(uiGroundStone, 70)},
		{142, 548, 2.4, alpha(uiGroundAsh, 58)},
		{236, 582, 1.9, alpha(uiGroundStone, 76)},
		{348, 534, 1.7, alpha(uiGroundDust, 60)},
		{462, 596, 2.1, alpha(uiGroundStone, 72)},
		{586, 548, 1.6, alpha(uiGroundAsh, 70)},
		{702, 588, 2.4, alpha(uiGroundDust, 62)},
		{816, 546, 2.0, alpha(uiGroundStone, 74)},
		{904, 606, 1.8, alpha(uiGroundAsh, 70)},
	}
	for _, stone := range stones {
		drawCircle(screen, stone.x, stone.y, stone.r, stone.c)
	}

	dustBands := []struct {
		x float64
		y float64
		w float64
		h float64
		c color.RGBA
	}{
		{48, 214, 72, 2, alpha(uiGroundDust, 18)},
		{164, 182, 54, 2, alpha(uiGroundAsh, 16)},
		{284, 332, 64, 2, alpha(uiGroundDust, 18)},
		{392, 248, 48, 2, alpha(uiGroundAsh, 16)},
		{518, 426, 68, 2, alpha(uiGroundDust, 18)},
		{648, 286, 56, 2, alpha(uiGroundAsh, 16)},
		{738, 248, 74, 2, alpha(uiGroundDust, 18)},
		{824, 516, 62, 2, alpha(uiGroundAsh, 18)},
	}
	for _, band := range dustBands {
		ebitenutil.DrawRect(screen, band.x, band.y, band.w, band.h, band.c)
	}
}

func drawWall(screen *ebiten.Image, w *wall) {
	if w.guard {
		drawGuardWall(screen, w)
		return
	}

	if w.destructive {
		rate := float64(w.hp) / float64(maxInt(w.maxHP, 1))
		if rate < 0 {
			rate = 0
		}
		base := blend(color.RGBA{106, 70, 52, 255}, color.RGBA{166, 108, 72, 255}, rate)
		shadow := shift(base, -26, -26, -26)
		highlight := shift(base, 30, 22, 12)
		ebitenutil.DrawRect(screen, w.box.x, w.box.y, w.box.w, w.box.h, shadow)
		ebitenutil.DrawRect(screen, w.box.x+1, w.box.y+1, maxF(w.box.w-2, 1), maxF(w.box.h-2, 1), base)
		ebitenutil.DrawRect(screen, w.box.x+2, w.box.y+2, maxF(w.box.w-4, 1), 2, alpha(highlight, 180))
		ebitenutil.DrawRect(screen, w.box.x+2, w.box.y+w.box.h-4, maxF(w.box.w-4, 1), 2, alpha(shadow, 220))
		brickH := 8.0
		brickW := 16.0
		for y := w.box.y + 1; y < w.box.y+w.box.h-1; y += brickH {
			ebitenutil.DrawLine(screen, w.box.x+1, y, w.box.x+w.box.w-1, y, color.RGBA{84, 62, 48, 150})
			row := int((y - w.box.y) / brickH)
			offset := 0.0
			if row%2 == 1 {
				offset = brickW / 2
			}
			for x := w.box.x + 3 - offset; x < w.box.x+w.box.w-2; x += brickW {
				if x > w.box.x+1 && x < w.box.x+w.box.w-2 {
					ebitenutil.DrawLine(screen, x, y, x, math.Min(y+brickH, w.box.y+w.box.h-1), color.RGBA{94, 68, 52, 124})
				}
			}
		}
		if rate < 0.75 {
			ebitenutil.DrawLine(screen, w.box.x+5, w.box.y+4, w.box.x+w.box.w*0.58, w.box.y+w.box.h-4, color.RGBA{56, 34, 28, 190})
		}
		if rate < 0.45 {
			ebitenutil.DrawLine(screen, w.box.x+w.box.w-6, w.box.y+5, w.box.x+w.box.w*0.36, w.box.y+w.box.h-4, color.RGBA{44, 26, 22, 210})
		}
		return
	}

	base := color.RGBA{94, 108, 122, 255}
	shadow := shift(base, -28, -28, -28)
	highlight := shift(base, 28, 28, 22)
	ebitenutil.DrawRect(screen, w.box.x, w.box.y, w.box.w, w.box.h, shadow)
	ebitenutil.DrawRect(screen, w.box.x+1, w.box.y+1, maxF(w.box.w-2, 1), maxF(w.box.h-2, 1), base)
	ebitenutil.DrawRect(screen, w.box.x+2, w.box.y+2, maxF(w.box.w-4, 1), 2, alpha(highlight, 170))
	ebitenutil.DrawRect(screen, w.box.x+2, w.box.y+w.box.h-4, maxF(w.box.w-4, 1), 2, alpha(shadow, 200))
	for x := w.box.x + 10; x < w.box.x+w.box.w-8; x += 18 {
		drawCircle(screen, x, w.box.y+w.box.h/2, 1.8, color.RGBA{196, 208, 218, 120})
	}
}

func drawGuardWall(screen *ebiten.Image, w *wall) {
	base := color.RGBA{132, 104, 74, 255}
	shadow := shift(base, -24, -24, -24)
	highlight := shift(base, 26, 18, 12)
	ebitenutil.DrawRect(screen, w.box.x, w.box.y, w.box.w, w.box.h, shadow)
	ebitenutil.DrawRect(screen, w.box.x+1, w.box.y+1, maxF(w.box.w-2, 1), maxF(w.box.h-2, 1), base)
	ebitenutil.DrawRect(screen, w.box.x+2, w.box.y+2, maxF(w.box.w-4, 1), 2, alpha(highlight, 150))
}

func drawFortress(screen *ebiten.Image, fort fortress) {
	cx := fort.box.x + fort.box.w/2
	cy := fort.box.y + fort.box.h/2
	rate := float64(fort.hp) / float64(maxInt(fort.maxHP, 1))
	if rate < 0 {
		rate = 0
	}
	glow := uiSignalGreen
	if rate < 0.45 {
		glow = uiSignalRed
	}
	drawCircle(screen, cx, cy, 54, alpha(glow, 22))
	ebitenutil.DrawRect(screen, fort.box.x-6, fort.box.y+fort.box.h-2, fort.box.w+12, 8, color.RGBA{20, 24, 28, 220})
	ebitenutil.DrawRect(screen, fort.box.x, fort.box.y, fort.box.w, fort.box.h, color.RGBA{54, 62, 74, 255})
	ebitenutil.DrawRect(screen, fort.box.x+3, fort.box.y+3, fort.box.w-6, fort.box.h-6, color.RGBA{102, 118, 134, 255})
	ebitenutil.DrawRect(screen, fort.box.x+10, fort.box.y+10, fort.box.w-20, fort.box.h-20, color.RGBA{36, 44, 54, 255})
	ebitenutil.DrawRect(screen, fort.box.x+8, fort.box.y+fort.box.h/2-3, fort.box.w-16, 6, color.RGBA{28, 34, 42, 230})
	drawCircle(screen, cx, cy, 6, alpha(glow, 190))
	drawRectOutline(screen, fort.box.x-2, fort.box.y-2, fort.box.w+4, fort.box.h+4, 1, alpha(glow, 120))
}

func drawPowerup(screen *ebiten.Image, p *powerup) {
	base := uiSteelBlue
	label := "S"
	switch p.kind {
	case powerRapid:
		base = uiSignalAmber
		label = "R"
	case powerRepair:
		base = uiSignalGreen
		label = "F"
	}
	pulseScale := pulse(p.life, 0.18, 0.78, 1.08)
	cx := p.box.x + p.box.w/2
	cy := p.box.y + p.box.h/2
	drawCircle(screen, cx, cy, 16*pulseScale, alpha(base, 24))
	ebitenutil.DrawRect(screen, p.box.x-2, p.box.y-2, p.box.w+4, p.box.h+4, alpha(uiInk, 200))
	ebitenutil.DrawRect(screen, p.box.x, p.box.y, p.box.w, p.box.h, base)
	drawRectOutline(screen, p.box.x-1, p.box.y-1, p.box.w+2, p.box.h+2, 1, alpha(base, 220))
	ebitenutil.DebugPrintAt(screen, label, int(p.box.x+4), int(p.box.y+3))
}

func drawBullet(screen *ebiten.Image, b *bullet) {
	core := color.RGBA{255, 244, 177, 255}
	glow := color.RGBA{255, 170, 64, 120}
	if !b.fromPlayer {
		core = color.RGBA{255, 149, 149, 255}
		glow = color.RGBA{255, 88, 88, 110}
	}
	trailX := b.x - b.vx*2.4
	trailY := b.y - b.vy*2.4
	ebitenutil.DrawLine(screen, b.x+bulletSize/2, b.y+bulletSize/2, trailX+bulletSize/2, trailY+bulletSize/2, alpha(glow, 160))
	ebitenutil.DrawRect(screen, b.x-2, b.y-2, bulletSize+4, bulletSize+4, glow)
	ebitenutil.DrawRect(screen, b.x, b.y, bulletSize, bulletSize, core)
}

func drawExplosion(screen *ebiten.Image, ex *explosion, frame int) {
	progress := 1 - float64(ex.life)/float64(maxInt(ex.max, 1))
	r := ex.radius * (0.35 + progress)
	alphaFade := uint8(float64(220) * (1 - progress))
	drawCircle(screen, ex.x, ex.y, r, color.RGBA{255, 160, 84, alphaFade / 2})
	drawCircle(screen, ex.x, ex.y, r*0.74, color.RGBA{255, 204, 128, alphaFade})
	drawCircle(screen, ex.x, ex.y, r*0.42, color.RGBA{255, 238, 182, alphaFade})
	if frame%4 < 2 {
		drawCircle(screen, ex.x, ex.y, r*1.08, color.RGBA{255, 110, 74, alphaFade / 3})
	}
}

func drawTank(screen *ebiten.Image, t tank, body, accent color.Color) {
	cx := t.x + tankSize/2
	cy := t.y + tankSize/2
	bodyRGBA := toRGBA(body)
	accentRGBA := toRGBA(accent)
	trackDark := color.RGBA{20, 26, 30, 255}
	trackLight := color.RGBA{66, 72, 78, 255}
	hullShadow := shift(bodyRGBA, -34, -34, -34)
	hullHighlight := shift(bodyRGBA, 26, 24, 24)
	turretShadow := shift(accentRGBA, -14, -14, -14)
	barrel := color.RGBA{214, 218, 224, 255}

	ebitenutil.DrawRect(screen, t.x+3, t.y+tankSize-2, tankSize-6, 4, color.RGBA{10, 12, 16, 120})
	ebitenutil.DrawRect(screen, t.x+1, t.y+2, 8, tankSize-4, trackDark)
	ebitenutil.DrawRect(screen, t.x+tankSize-9, t.y+2, 8, tankSize-4, trackDark)
	for i := 0; i < 4; i++ {
		wy := t.y + 6 + float64(i)*7
		drawCircle(screen, t.x+5, wy, 1.7, trackLight)
		drawCircle(screen, t.x+tankSize-5, wy, 1.7, trackLight)
	}
	ebitenutil.DrawRect(screen, t.x+7, t.y+5, tankSize-14, tankSize-10, hullShadow)
	ebitenutil.DrawRect(screen, t.x+8, t.y+6, tankSize-16, tankSize-12, bodyRGBA)
	ebitenutil.DrawRect(screen, t.x+10, t.y+8, tankSize-20, 4, alpha(hullHighlight, 180))
	ebitenutil.DrawRect(screen, t.x+12, t.y+tankSize-10, tankSize-24, 3, alpha(hullShadow, 200))

	drawCircle(screen, cx, cy, 9.2, turretShadow)
	drawCircle(screen, cx, cy, 7.2, accentRGBA)
	drawCircle(screen, cx, cy, 3, color.RGBA{68, 72, 80, 255})

	switch t.turret {
	case up:
		ebitenutil.DrawRect(screen, cx-2.5, t.y-11, 5, 16, barrel)
		ebitenutil.DrawRect(screen, cx-4, t.y-13, 8, 3, color.RGBA{142, 148, 156, 255})
		ebitenutil.DrawRect(screen, cx-1, t.y-13, 2, 3, accentRGBA)
	case down:
		ebitenutil.DrawRect(screen, cx-2.5, t.y+tankSize-5, 5, 16, barrel)
		ebitenutil.DrawRect(screen, cx-4, t.y+tankSize+10, 8, 3, color.RGBA{142, 148, 156, 255})
		ebitenutil.DrawRect(screen, cx-1, t.y+tankSize+10, 2, 3, accentRGBA)
	case left:
		ebitenutil.DrawRect(screen, t.x-11, cy-2.5, 16, 5, barrel)
		ebitenutil.DrawRect(screen, t.x-13, cy-4, 3, 8, color.RGBA{142, 148, 156, 255})
		ebitenutil.DrawRect(screen, t.x-13, cy-1, 3, 2, accentRGBA)
	case right:
		ebitenutil.DrawRect(screen, t.x+tankSize-5, cy-2.5, 16, 5, barrel)
		ebitenutil.DrawRect(screen, t.x+tankSize+10, cy-4, 3, 8, color.RGBA{142, 148, 156, 255})
		ebitenutil.DrawRect(screen, t.x+tankSize+10, cy-1, 3, 2, accentRGBA)
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

func hudTextWidth(s string) int {
	return len([]rune(s)) * 6
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

func hudBottomY() int {
	return hudTopY + hudHeight
}

func messageBoxTopY() int {
	return hudBottomY() + hudMessageGap
}

func menuOptionsBottomY(itemCount int) int {
	return computeMenuLayout(itemCount).optionsBottomY
}

func menuFooterTopY() int {
	return menuFooterY
}

func menuOptionsTopY() int {
	return computeMenuLayout(menuItemCount).optionsTopY
}

func menuHelpSectionBottomY() int {
	return computeMenuLayout(menuItemCount).helpBottomY
}

func menuHelpToOptionsDistanceY() int {
	return menuOptionsTopY() - menuHelpSectionBottomY()
}

func menuTitleX() int {
	return centeredTextX(menuTitleText, menuHeaderX, menuHeaderW)
}

func menuTitleY() int {
	return centeredTextY(menuHeaderY, menuHeaderH, menuTextHeight)
}

func menuHelpTextX(s string) int {
	return centeredTextX(s, 110, 484)
}

type menuLayout struct {
	helpLineY      [menuHelpLineCount]int
	helpBottomY    int
	optionsTopY    int
	optionsBottomY int
	optionBoxTopY  []int
	optionTextY    []int
}

func computeMenuLayout(optionCount int) menuLayout {
	l := menuLayout{
		optionBoxTopY: make([]int, 0, maxInt(optionCount, 0)),
		optionTextY:   make([]int, 0, maxInt(optionCount, 0)),
	}

	helpStartY := menuHeaderY + menuHeaderH + 2
	for i := 0; i < menuHelpLineCount; i++ {
		l.helpLineY[i] = helpStartY + i*menuHelpLineGap
	}
	l.helpBottomY = menuHeaderY + menuHeaderH + 2
	l.optionsTopY = l.helpBottomY + menuHelpToOptionGap

	if optionCount <= 0 {
		l.optionsBottomY = l.optionsTopY
		return l
	}

	gap := menuOptionMinGap
	if optionCount > 1 {
		available := menuFooterY - menuOptionBottomPadding - l.optionsTopY
		totalBoxHeight := optionCount * menuOptionBoxHeight
		minNeed := totalBoxHeight + (optionCount-1)*menuOptionMinGap
		if available > minNeed {
			flexible := (available - totalBoxHeight) / (optionCount - 1)
			gap = clampInt(flexible, menuOptionMinGap, menuOptionMaxGap)
		}
	}

	for i := 0; i < optionCount; i++ {
		top := l.optionsTopY + i*(menuOptionBoxHeight+gap)
		l.optionBoxTopY = append(l.optionBoxTopY, top)
		l.optionTextY = append(l.optionTextY, top+menuOptionTextTopInset)
	}
	l.optionsBottomY = l.optionBoxTopY[len(l.optionBoxTopY)-1] + menuOptionBoxHeight
	return l
}

func centeredTextX(s string, areaX, areaW int) int {
	return areaX + (areaW-textWidth(s))/2
}

func centeredTextY(areaY, areaH, textH int) int {
	return areaY + (areaH-textH)/2
}

func drawStatusPanel(screen *ebiten.Image, panelW, panelH int, accent color.RGBA, lines ...string) {
	if len(lines) == 0 {
		return
	}
	x := float64(screenW/2 - panelW/2)
	y := float64(screenH/2 - panelH/2)
	drawSurfacePanel(screen, x, y, float64(panelW), float64(panelH), accent)

	const lineGap = 10
	textBlockH := len(lines)*menuTextHeight + (len(lines)-1)*lineGap
	startY := int(y) + (panelH-textBlockH)/2
	for i, line := range lines {
		lineY := startY + i*(menuTextHeight+lineGap)
		ebitenutil.DebugPrintAt(screen, line, centeredTextX(line, int(x), panelW), lineY)
	}
}

func drawPillLabel(screen *ebiten.Image, x, y, w, h float64, accent color.RGBA, text string) {
	drawPill(screen, x, y, w, h, accent)
	ebitenutil.DebugPrintAt(screen, text, centeredTextX(text, int(x), int(w)), centeredTextY(int(y), int(h), menuTextHeight))
}
