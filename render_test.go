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
