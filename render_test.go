package tankbattle

import (
	"strings"
	"testing"
)

func TestPlayerEnergySummaryMergedRightValue(t *testing.T) {
	p := tank{hp: 4, maxHP: 5, turretHP: 3, turretMaxHP: 6}
	s := playerEnergySummary(p)
	if !strings.Contains(s, "H:4/5") {
		t.Fatalf("hull summary missing: %q", s)
	}
	if !strings.Contains(s, "T:3/6") {
		t.Fatalf("turret summary missing: %q", s)
	}
}

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
