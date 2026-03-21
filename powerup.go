package tankbattle

import "math/rand"

func (g *game) updatePowerups() {
	keep := g.powerups[:0]
	pr := tankRect(g.player)
	for _, p := range g.powerups {
		p.life--
		if p.life <= 0 {
			continue
		}
		if overlap(pr, p.box) {
			switch p.kind {
			case powerShield:
				g.shieldTick = 600
				g.setMessage("Shield online", 90)
			case powerRapid:
				g.rapidTick = 600
				g.setMessage("Rapid fire", 90)
			case powerRepair:
				g.fort.hp += 4
				if g.fort.hp > g.fort.maxHP {
					g.fort.hp = g.fort.maxHP
				}
				g.player.hp += 1
				if g.player.hp > g.player.maxHP {
					g.player.hp = g.player.maxHP
				}
				g.player.turretHP += 2
				if g.player.turretHP > g.player.turretMaxHP {
					g.player.turretHP = g.player.turretMaxHP
				}
				g.setMessage("Fortress & tank repaired", 90)
			}
			g.score += 80
			g.playSFX(sfxPowerupPickup)
			continue
		}
		keep = append(keep, p)
	}
	g.powerups = keep
}

func (g *game) trySpawnRandomPowerup() {
	if g.frame%420 != 0 || len(g.powerups) >= powerupMaxActive {
		return
	}
	x := float64(rand.Intn(screenW-180) + 90)
	y := float64(rand.Intn(screenH-300) + 90)
	box := rect{x: x, y: y, w: powerupSize, h: powerupSize}
	if !g.canPlacePowerup(box) {
		return
	}
	kind := powerupKind(rand.Intn(3))
	g.powerups = append(g.powerups, &powerup{kind: kind, box: box, life: 1200})
	g.playSFX(sfxPowerupSpawn)
}

func (g *game) dropPowerup(x, y float64) {
	if len(g.powerups) >= powerupMaxActive {
		return
	}
	kindRoll := rand.Intn(100)
	kind := powerRapid
	if kindRoll < 34 {
		kind = powerShield
	} else if kindRoll < 67 {
		kind = powerRepair
	}
	g.powerups = append(g.powerups, &powerup{kind: kind, box: rect{x: x - powerupSize/2, y: y - powerupSize/2, w: powerupSize, h: powerupSize}, life: 900})
	g.playSFX(sfxPowerupSpawn)
}

func (g *game) canPlacePowerup(box rect) bool {
	if overlap(box, g.fort.box) || overlap(box, tankRect(g.player)) {
		return false
	}
	for _, w := range g.walls {
		if overlap(box, w.box) {
			return false
		}
	}
	for _, e := range g.enemies {
		if overlap(box, tankRect(*e)) {
			return false
		}
	}
	for _, p := range g.powerups {
		if overlap(box, p.box) {
			return false
		}
	}
	return true
}
