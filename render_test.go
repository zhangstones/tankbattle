package tankbattle

import (
	"strings"
	"testing"
	"time"
)

func TestPlayerCombinedEnergy(t *testing.T) {
	p := tank{hp: 4, maxHP: 5, turretHP: 3, turretMaxHP: 6}
	now, max := playerCombinedEnergy(p)
	if now != 7 || max != 11 {
		t.Fatalf("combined tank energy mismatch: now=%d max=%d", now, max)
	}
}

func TestMessageBoxAvoidsHUD(t *testing.T) {
	if messageBoxTopY() <= hudBottomY() {
		t.Fatalf("message box should be below HUD: msgTop=%d hudBottom=%d", messageBoxTopY(), hudBottomY())
	}
}

func TestHUDAndStatusUseThinFrame(t *testing.T) {
	if hudFrameInset > 2 {
		t.Fatalf("HUD frame should be thin, got inset=%d", hudFrameInset)
	}
	if statusInset > 2 {
		t.Fatalf("status frame should be thin, got inset=%d", statusInset)
	}
}

func TestMenuOptionsDoNotOverlapFooterTipBar(t *testing.T) {
	if menuOptionsBottomY(menuItemCount) >= menuFooterTopY() {
		t.Fatalf("menu options should stay above tip bar: optionsBottom=%d footerTop=%d", menuOptionsBottomY(menuItemCount), menuFooterTopY())
	}
}

func TestMenuOptionsKeepTopSpacingFromHelpSection(t *testing.T) {
	if menuOptionsTopY() <= menuHelpSectionBottomY() {
		t.Fatalf("menu options should be below help section: optionsTop=%d helpBottom=%d", menuOptionsTopY(), menuHelpSectionBottomY())
	}
}

func TestMenuHelpAndOptionsDistanceIsCompact(t *testing.T) {
	gap := menuHelpToOptionsDistanceY()
	if gap < 8 || gap > 20 {
		t.Fatalf("menu help/options gap should be compact and readable: got %d", gap)
	}
}

func TestMenuTitleIsCenteredInHeader(t *testing.T) {
	leftMargin := menuTitleX() - menuHeaderX
	rightMargin := (menuHeaderX + menuHeaderW) - (menuTitleX() + textWidth(menuTitleText))
	diff := leftMargin - rightMargin
	if diff < 0 {
		diff = -diff
	}
	if diff > 1 {
		t.Fatalf("menu title should be centered: leftMargin=%d rightMargin=%d", leftMargin, rightMargin)
	}
}

func TestMenuTitleIsVerticallyCenteredInHeader(t *testing.T) {
	topMargin := menuTitleY() - menuHeaderY
	bottomMargin := (menuHeaderY + menuHeaderH) - (menuTitleY() + menuTextHeight)
	diff := topMargin - bottomMargin
	if diff < 0 {
		diff = -diff
	}
	if diff > 1 {
		t.Fatalf("menu title should be vertically centered: topMargin=%d bottomMargin=%d", topMargin, bottomMargin)
	}
}

func TestFireHelpTextMergedWithMainHelpLine(t *testing.T) {
	if !strings.Contains(menuHelpLine3, "FIRE J/Space") {
		t.Fatalf("fire hint should be merged into combat help line: %q", menuHelpLine3)
	}
}

func TestHistoryPanelLayoutIsSpacious(t *testing.T) {
	if historyPanelW < 700 {
		t.Fatalf("history panel should be wide enough, got width=%d", historyPanelW)
	}
	if hudRankingLineGap < 22 {
		t.Fatalf("history row line gap should be spacious, got gap=%d", hudRankingLineGap)
	}
}

func TestMenuPanelsStayInsideScreen(t *testing.T) {
	if menuPanelX < 0 || menuPanelY < 0 {
		t.Fatalf("menu panel origin should stay on screen: x=%d y=%d", menuPanelX, menuPanelY)
	}
	if menuPanelX+menuPanelW > screenW {
		t.Fatalf("menu panel should fit screen width: right=%d screen=%d", menuPanelX+menuPanelW, screenW)
	}
	if menuPanelY+menuPanelH > screenH {
		t.Fatalf("menu panel should fit screen height: bottom=%d screen=%d", menuPanelY+menuPanelH, screenH)
	}
}

func TestHUDPanelsFitScreenWidth(t *testing.T) {
	leftPanelRight := 12 + 556
	rightPanelRight := 582 + 366
	if leftPanelRight >= 582 {
		t.Fatalf("hud left and right panels should keep a gap: leftRight=%d rightX=%d", leftPanelRight, 582)
	}
	if rightPanelRight > screenW {
		t.Fatalf("hud right panel should fit screen width: right=%d screen=%d", rightPanelRight, screenW)
	}
}

func TestHistoryPanelFitsWithinScreen(t *testing.T) {
	if historyPanelX+historyPanelW > screenW {
		t.Fatalf("history panel should fit screen width: right=%d screen=%d", historyPanelX+historyPanelW, screenW)
	}
	if historyPanelY+historyPanelH > screenH {
		t.Fatalf("history panel should fit screen height: bottom=%d screen=%d", historyPanelY+historyPanelH, screenH)
	}
}

func TestMenuSidebarAndDescriptionsStayWithinTextColumns(t *testing.T) {
	const (
		menuOptionsX        = 92
		menuOptionsW        = 520
		menuValuePillW      = 128
		menuValueColumnGap  = 16
		menuDescTextX       = 116
		menuSidebarX        = 640
		menuSidebarTextX    = 658
		menuSidebarPanelW   = 232
		menuSidebarRightPad = 18
	)
	menuValuePillX := menuOptionsX + menuOptionsW - menuValuePillW - menuValueColumnGap
	menuDescMaxWidth := menuValuePillX - menuDescTextX - 16
	for _, desc := range []string{
		"Slower enemies and lighter mission pressure.",
		"Balanced pace and enemy fire cadence.",
		"Faster waves, higher HP, tighter pressure.",
	} {
		if textWidth(desc) > menuDescMaxWidth {
			t.Fatalf("menu description should fit value column gap: %q width=%d limit=%d", desc, textWidth(desc), menuDescMaxWidth)
		}
	}
	sidebarMaxWidth := (menuSidebarX + 16 + menuSidebarPanelW) - menuSidebarTextX - menuSidebarRightPad
	for _, line := range []string{
		"Mode: NORMAL",
		"Audio: ON 75%",
		"Up/Down: browse",
		"Left/Right: adjust",
		"Enter: start",
		"M: back or resume",
		"1/2/3 set difficulty.",
		"R restarts in battle.",
		"H opens score history.",
		"Setup changes restart run.",
		"Audio changes keep run.",
	} {
		if textWidth(line) > sidebarMaxWidth {
			t.Fatalf("sidebar text should fit sidebar card: %q width=%d limit=%d", line, textWidth(line), sidebarMaxWidth)
		}
	}
}

func TestRightHUDValuesDoNotOverlapThreatBox(t *testing.T) {
	const (
		rightHUDX      = 582
		rightHUDW      = 366
		vitalValueX    = 210
		threatBoxW     = 82
		threatInsetX   = 18
		maxVitalWidth  = len("10/10") * 7
	)
	threatX := rightHUDX + rightHUDW - threatBoxW - threatInsetX
	vitalValueRight := rightHUDX + vitalValueX + maxVitalWidth
	if vitalValueRight >= threatX {
		t.Fatalf("right hud values should stay left of threat box: valueRight=%d threatX=%d", vitalValueRight, threatX)
	}
}

func TestFormatDuration(t *testing.T) {
	if got := formatDuration(0); got != "00:00" {
		t.Fatalf("duration format mismatch: got %q", got)
	}
	if got := formatDuration(125); got != "02:05" {
		t.Fatalf("duration format mismatch: got %q", got)
	}
}

func TestFormatScoreTimeUsesLocalClock(t *testing.T) {
	ts := time.Date(2026, 3, 22, 10, 0, 0, 0, time.UTC).Format(time.RFC3339)
	got := formatScoreTime(ts)
	if !strings.Contains(got, "2026-03-22") {
		t.Fatalf("formatted time should include date, got %q", got)
	}
	if len(got) != len("2006-01-02 15:04:05") {
		t.Fatalf("formatted time should have full local timestamp, got %q", got)
	}
}
