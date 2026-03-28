package tankbattle

import "testing"

func TestUpdatePowerupsShieldPickup(t *testing.T) {
	g := newPlayingGameForTest()
	mock := &mockSFXPlayer{enabled: true}
	g.audio = mock
	g.powerups = []*powerup{{kind: powerShield, box: rect{g.player.x, g.player.y, 16, 16}, life: 100}}
	g.updatePowerups()
	if g.shieldTick == 0 {
		t.Fatalf("shield powerup should apply")
	}
	if last, ok := mock.last(); !ok || last != sfxPowerupPickupShield {
		t.Fatalf("expected shield pickup sfx, got %v (ok=%v)", last, ok)
	}
}

func TestUpdatePowerupsRapidPickup(t *testing.T) {
	g := newPlayingGameForTest()
	mock := &mockSFXPlayer{enabled: true}
	g.audio = mock
	g.powerups = []*powerup{{kind: powerRapid, box: rect{g.player.x, g.player.y, 16, 16}, life: 100}}
	g.updatePowerups()
	if g.rapidTick == 0 {
		t.Fatalf("rapid powerup should apply")
	}
	if last, ok := mock.last(); !ok || last != sfxPowerupPickupRapid {
		t.Fatalf("expected rapid pickup sfx, got %v (ok=%v)", last, ok)
	}
}

func TestUpdatePowerupsRepairCapped(t *testing.T) {
	g := newPlayingGameForTest()
	mock := &mockSFXPlayer{enabled: true}
	g.audio = mock
	g.fort.hp = g.fort.maxHP - 1
	g.player.hp = g.player.maxHP - 1
	g.player.turretHP = g.player.turretMaxHP - 1
	g.powerups = []*powerup{{kind: powerRepair, box: rect{g.player.x, g.player.y, 16, 16}, life: 100}}
	g.updatePowerups()
	if g.fort.hp != g.fort.maxHP {
		t.Fatalf("repair should cap at max hp")
	}
	if g.player.hp != g.player.maxHP {
		t.Fatalf("repair should cap player hp at max")
	}
	if g.player.turretHP != g.player.turretMaxHP {
		t.Fatalf("repair should cap player turret hp at max")
	}
	if last, ok := mock.last(); !ok || last != sfxPowerupPickupRepair {
		t.Fatalf("expected repair pickup sfx, got %v (ok=%v)", last, ok)
	}
}

func TestUpdatePowerupsExpire(t *testing.T) {
	g := newPlayingGameForTest()
	g.powerups = []*powerup{{kind: powerRepair, box: rect{10, 10, 16, 16}, life: 1}}
	g.updatePowerups()
	if len(g.powerups) != 0 {
		t.Fatalf("expired powerup should be removed")
	}
}

func TestTrySpawnRandomPowerupFrameGate(t *testing.T) {
	g := newPlayingGameForTest()
	g.frame = 1
	g.powerups = nil
	g.trySpawnRandomPowerup()
	if len(g.powerups) != 0 {
		t.Fatalf("powerup should not spawn when frame%%420!=0")
	}
}

func TestDropPowerupRespectsCap(t *testing.T) {
	g := newPlayingGameForTest()
	g.powerups = []*powerup{{}, {}, {}}
	g.dropPowerup(100, 100)
	if len(g.powerups) != powerupMaxActive {
		t.Fatalf("drop should respect max cap")
	}
}

func TestCanPlacePowerupRejectsActorOverlap(t *testing.T) {
	g := newPlayingGameForTest()
	g.enemies = []*tank{{x: 200, y: 200}}

	if g.canPlacePowerup(rect{x: g.player.x, y: g.player.y, w: powerupSize, h: powerupSize}) {
		t.Fatalf("should reject powerup overlapping player")
	}
	if g.canPlacePowerup(rect{x: 200, y: 200, w: powerupSize, h: powerupSize}) {
		t.Fatalf("should reject powerup overlapping enemy")
	}
}

func TestCanPlacePowerupRejectsExistingPowerupOverlap(t *testing.T) {
	g := newPlayingGameForTest()
	g.powerups = []*powerup{{kind: powerShield, box: rect{x: 120, y: 140, w: powerupSize, h: powerupSize}, life: 100}}
	if g.canPlacePowerup(rect{x: 124, y: 144, w: powerupSize, h: powerupSize}) {
		t.Fatalf("should reject overlap with existing powerup")
	}
}
