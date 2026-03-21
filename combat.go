package tankbattle

import "math/rand"

func (g *game) updateBullets() {
	for _, b := range g.bullets {
		if !b.alive {
			continue
		}
		b.x += b.vx
		b.y += b.vy
		if b.x < 0 || b.y < 0 || b.x > screenW-bulletSize || b.y > screenH-bulletSize {
			b.alive = false
			continue
		}
		br := rect{b.x, b.y, bulletSize, bulletSize}

		for _, w := range g.walls {
			if overlap(br, w.box) {
				b.alive = false
				if w.destructive {
					w.hp -= b.dmg
				}
				if !b.fromPlayer && w.guard {
					g.score = maxInt(0, g.score-guardHitLoss)
				}
				g.spawnExplosion(b.x, b.y, 14)
				g.playSFX(sfxHitWall)
				g.playSFX(sfxExplosionSmall)
				break
			}
		}
		if !b.alive {
			continue
		}

		if overlap(br, g.fort.box) {
			if !b.fromPlayer {
				g.fort.hp -= fortHitDamage
				if g.fort.hp < 0 {
					g.fort.hp = 0
				}
				g.score = maxInt(0, g.score-fortHitLoss)
			}
			b.alive = false
			g.spawnExplosion(b.x, b.y, 22)
			g.playSFX(sfxHitFortress)
			g.playSFX(sfxExplosionLarge)
			continue
		}

		if b.fromPlayer {
			for i := 0; i < len(g.enemies); i++ {
				e := g.enemies[i]
				if overlap(br, tankRect(*e)) {
					b.alive = false
					e.hp -= b.dmg
					g.spawnExplosion(b.x, b.y, 18)
					g.playSFX(sfxHitTank)
					g.playSFX(sfxExplosionSmall)
					if e.hp <= 0 {
						g.spawnExplosion(e.x+tankSize/2, e.y+tankSize/2, 30)
						g.playSFX(sfxExplosionLarge)
						if rand.Intn(100) < 30 {
							g.dropPowerup(e.x+tankSize/2, e.y+tankSize/2)
						}
						g.enemies = append(g.enemies[:i], g.enemies[i+1:]...)
						g.score += 150
						g.enemyKills++
						i--
					}
					break
				}
			}
		} else if overlap(br, tankRect(g.player)) {
			b.alive = false
			if g.shieldTick > 0 {
				g.spawnExplosion(g.player.x+tankSize/2, g.player.y+tankSize/2, 12)
				g.playSFX(sfxHitWall)
			} else {
				g.player.hp -= b.dmg
				g.player.turretHP -= b.dmg
				if g.player.turretHP < 0 {
					g.player.turretHP = 0
				}
				g.spawnExplosion(b.x, b.y, 20)
				g.playSFX(sfxHitTank)
				g.playSFX(sfxExplosionSmall)
			}
		}
	}

	alive := g.bullets[:0]
	for _, b := range g.bullets {
		if b.alive {
			alive = append(alive, b)
		}
	}
	g.bullets = alive
}

func (g *game) fire(t *tank, fromPlayer bool) {
	speed := 6.8
	if !fromPlayer {
		speed = 5.6
	}
	bx := t.x + tankSize/2 - bulletSize/2
	by := t.y + tankSize/2 - bulletSize/2
	vx, vy := 0.0, 0.0
	fireDir := t.turret
	if fireDir != up && fireDir != down && fireDir != left && fireDir != right {
		fireDir = t.dir
	}
	switch fireDir {
	case up:
		vy = -speed
		by = t.y - bulletSize
	case down:
		vy = speed
		by = t.y + tankSize
	case left:
		vx = -speed
		bx = t.x - bulletSize
	case right:
		vx = speed
		bx = t.x + tankSize
	}
	g.bullets = append(g.bullets, &bullet{x: bx, y: by, vx: vx, vy: vy, fromPlayer: fromPlayer, alive: true, dmg: 1})
	if fromPlayer {
		g.playSFX(sfxShootPlayer)
	} else {
		g.playSFX(sfxShootEnemy)
	}
}
