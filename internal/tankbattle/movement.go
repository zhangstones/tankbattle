package tankbattle

import (
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *game) updatePlayer() {
	if g.player.cooldown > 0 {
		g.player.cooldown--
	}
	g.handlePlayerTurnInput()

	dx, dy := 0.0, 0.0
	w := ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp)
	s := ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown)
	a := ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft)
	d := ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight)
	if g.canMoveOnHeld(up, w) {
		dy = -g.player.speed
	}
	if g.canMoveOnHeld(down, s) {
		dy = g.player.speed
	}
	if g.canMoveOnHeld(left, a) {
		dx = -g.player.speed
	}
	if g.canMoveOnHeld(right, d) {
		dx = g.player.speed
	}
	g.tryMoveTank(&g.player, dx, dy)

	if (inpututil.IsKeyJustPressed(ebiten.KeyJ) || inpututil.IsKeyJustPressed(ebiten.KeySpace)) && g.player.cooldown == 0 {
		g.onPlayerFired()
		g.fire(&g.player, true)
		g.player.cooldown = playerFireCooldown(g.rapidTick > 0)
	}
}

func playerFireCooldown(rapid bool) int {
	if rapid {
		return playerRapidFireCooldownFrames
	}
	return playerFireCooldownFrames
}

func (g *game) handlePlayerTurnInput() {
	upTap := inpututil.IsKeyJustPressed(ebiten.KeyW) || inpututil.IsKeyJustPressed(ebiten.KeyArrowUp)
	downTap := inpututil.IsKeyJustPressed(ebiten.KeyS) || inpututil.IsKeyJustPressed(ebiten.KeyArrowDown)
	leftTap := inpututil.IsKeyJustPressed(ebiten.KeyA) || inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft)
	rightTap := inpututil.IsKeyJustPressed(ebiten.KeyD) || inpututil.IsKeyJustPressed(ebiten.KeyArrowRight)
	g.applyPlayerTurnTaps(upTap, downTap, leftTap, rightTap)
}

func (g *game) applyPlayerTurnTaps(upTap, downTap, leftTap, rightTap bool) {
	if upTap {
		g.onPlayerDirTap(up)
	}
	if downTap {
		g.onPlayerDirTap(down)
	}
	if leftTap {
		g.onPlayerDirTap(left)
	}
	if rightTap {
		g.onPlayerDirTap(right)
	}
}

func (g *game) onPlayerDirTap(d direction) bool {
	last := g.playerTapFrame[d]
	g.playerTapFrame[d] = g.frame
	if g.frame-last > playerTurnDoubleTapFrames {
		return false
	}
	g.player.dir = d
	g.player.turret = d
	g.playerMoveLockUntil = g.frame + playerTurnMoveLockFrames
	g.playerPressStart[d] = g.frame
	return true
}

func (g *game) canMoveOnHeld(d direction, pressed bool) bool {
	if !pressed {
		g.playerPressStart[d] = -9999
		return false
	}
	if inpututil.IsKeyJustPressed(keyFromDirection(d)) {
		g.playerPressStart[d] = g.frame
	}
	if g.playerPressStart[d] < -9000 {
		g.playerPressStart[d] = g.frame
	}
	if g.frame < g.playerMoveLockUntil {
		return false
	}
	return g.frame-g.playerPressStart[d] >= playerTapGraceFrames
}

func keyFromDirection(d direction) ebiten.Key {
	switch d {
	case up:
		return ebiten.KeyW
	case down:
		return ebiten.KeyS
	case left:
		return ebiten.KeyA
	default:
		return ebiten.KeyD
	}
}

func (g *game) updateEnemies() {
	for _, e := range g.enemies {
		ensureEnemyTraits(e)
		e.age++
		if e.cooldown > 0 {
			e.cooldown--
		}
		if e.turnLock > 0 {
			e.turnLock--
		}
		e.aiTick--
		if e.aiTick <= 0 || e.stuck > 0 {
			targetDir := g.nextEnemyPlanDir(e)
			g.applyEnemyTurn(e, targetDir)
			e.aiTick = rand.Intn(4) + e.replan
		}
		dx, dy := nextStep(0, 0, e.dir, e.speed)
		if g.tryMoveTank(e, dx, dy) {
			e.stuck = 0
		} else {
			e.stuck++
			if g.tryUnstuck(e) {
				e.stuck = 0
			}
		}

		if e.cooldown == 0 {
			alignedBase := math.Abs((e.x+tankSize/2)-(g.fort.box.x+g.fort.box.w/2)) < 20 || math.Abs((e.y+tankSize/2)-(g.fort.box.y+g.fort.box.h/2)) < 20
			alignedPlayer := math.Abs((e.x+tankSize/2)-(g.player.x+tankSize/2)) < 20 || math.Abs((e.y+tankSize/2)-(g.player.y+tankSize/2)) < 20
			fireChance := enemyFireChance(alignedBase, alignedPlayer, g.enemyFireBonus()+e.fireBias, e.aiRand)
			fortressHitProb := g.enemyFortressHitProbability(e, alignedBase)
			chance := fireChance
			if fortressHitProb > chance {
				chance = fortressHitProb
			}
			if rand.Intn(100) < chance {
				e.turret = e.dir
				if rand.Intn(100) < fortressHitProb {
					e.turret = g.directionTowardFortress(e)
				}
				g.fire(e, false)
				e.cooldown = enemyFireCooldown(alignedBase, alignedPlayer, e.aiRand)
			}
		}
	}
}

func (g *game) nextEnemyPlanDir(e *tank) direction {
	target := g.flowDirForEnemy(e)
	if rand.Intn(100) >= 15 {
		return target
	}

	alts := []direction{left, right, up, down}
	rand.Shuffle(len(alts), func(i, j int) {
		alts[i], alts[j] = alts[j], alts[i]
	})
	for _, d := range alts {
		if d == oppositeDir(e.dir) {
			continue
		}
		sx, sy := nextStep(e.x, e.y, d, e.speed)
		cand := rect{x: sx, y: sy, w: tankSize, h: tankSize}
		if g.canOccupy(cand, e) {
			return d
		}
	}
	return target
}

func enemyFireChance(alignedBase, alignedPlayer bool, fireBonus int, randFactor float64) int {
	chance := 2 + fireBonus/4
	chance += int((randFactor - 0.5) * 8)
	if alignedBase {
		chance += 18
	}
	if alignedPlayer {
		chance += 12
	}
	if chance < 1 {
		return 1
	}
	if chance > 70 {
		return 70
	}
	return chance
}

func enemyFireCooldown(alignedBase, alignedPlayer bool, randFactor float64) int {
	cooldown := rand.Intn(enemyFireCooldownBaseVar) + enemyFireCooldownBaseMin
	cooldown += int((0.5 - randFactor) * 8)
	if alignedBase || alignedPlayer {
		cooldown -= rand.Intn(8)
	}
	if cooldown < enemyFireCooldownMin {
		return enemyFireCooldownMin
	}
	return cooldown
}

func ensureEnemyTraits(e *tank) {
	if e.replan <= 0 {
		e.replan = 8
	}
	if e.aggro <= 0 {
		e.aggro = 1
	}
}

func (g *game) onPlayerFired() {
	g.playerSilentFrames = 0
}

func (g *game) applyEnemyTurn(e *tank, target direction) {
	if target == e.dir {
		e.turnVote = 0
		return
	}
	if target != e.turnWant {
		e.turnWant = target
		e.turnVote = 1
	} else {
		e.turnVote++
	}

	// Prevent left-right / up-down jitter by requiring stronger evidence for full reverse turns.
	needVotes := 3
	lockFrames := 6
	if target == oppositeDir(e.dir) && e.stuck <= 1 {
		needVotes = 7
		lockFrames = 14
	}
	if e.stuck > 1 {
		needVotes = 1
		lockFrames = 4
	}
	if e.turnLock > 0 && e.stuck <= 1 {
		return
	}
	if e.turnVote >= needVotes {
		e.dir = target
		e.turnLock = lockFrames
		e.turnVote = 0
	}
}

func (g *game) tryUnstuck(e *tank) bool {
	cands := []direction{left, right, up, down}
	if e.dir == left || e.dir == right {
		cands = []direction{up, down, e.dir, oppositeDir(e.dir)}
	} else {
		cands = []direction{left, right, e.dir, oppositeDir(e.dir)}
	}
	for _, d := range cands {
		dx, dy := nextStep(0, 0, d, e.speed*0.9)
		if g.tryMoveTank(e, dx, dy) {
			e.dir = d
			e.turnLock = 8
			e.turnVote = 0
			e.turnWant = d
			return true
		}
	}
	return false
}

func (g *game) flowDirForEnemy(e *tank) direction {
	cx := e.x + tankSize/2
	cy := e.y + tankSize/2
	targetX, targetY := g.enemyTargetPoint(e, cx, cy)
	pressure := g.combinedFortressPressure(e)
	pull := 0.028 + pressure*0.022
	fx := (targetX - cx) * pull
	fy := (targetY - cy) * pull

	px := g.player.x + tankSize/2
	py := g.player.y + tankSize/2
	dx := cx - px
	dy := cy - py
	dist := math.Hypot(dx, dy)
	if dist > 0 && dist < 95 {
		scale := (95 - dist) / 95 * 2.6
		fx += dx / dist * scale
		fy += dy / dist * scale
	}

	for _, other := range g.enemies {
		if other == e {
			continue
		}
		ox := other.x + tankSize/2
		oy := other.y + tankSize/2
		ddx := cx - ox
		ddy := cy - oy
		sep := math.Hypot(ddx, ddy)
		if sep > 0 && sep < 96 {
			scale := (96 - sep) / 96 * 2.8
			fx += ddx / sep * scale
			fy += ddy / sep * scale
		}
	}

	for _, w := range g.walls {
		nx := clampF(cx, w.box.x, w.box.x+w.box.w)
		ny := clampF(cy, w.box.y, w.box.y+w.box.h)
		ddx := cx - nx
		ddy := cy - ny
		sep := math.Hypot(ddx, ddy)
		if sep > 0 && sep < 74 {
			scale := (74 - sep) / 74 * 2.5
			fx += ddx / sep * scale
			fy += ddy / sep * scale
		}
	}

	if cx < 64 {
		fx += (64 - cx) / 64 * 2.0
	}
	if cx > screenW-64 {
		fx -= (cx - (screenW - 64)) / 64 * 2.0
	}
	if cy < 64 {
		fy += (64 - cy) / 64 * 2.0
	}
	if cy > screenH-64 {
		fy -= (cy - (screenH - 64)) / 64 * 2.0
	}

	best := e.dir
	bestScore := -math.MaxFloat64
	currScore := -math.MaxFloat64
	baseX := g.fort.box.x + g.fort.box.w/2
	baseY := g.fort.box.y + g.fort.box.h/2
	linePressure := playerSilencePressure(g.playerSilentFrames)
	for _, d := range []direction{up, down, left, right} {
		sx, sy := nextStep(e.x, e.y, d, e.speed)
		cand := rect{x: sx, y: sy, w: tankSize, h: tankSize}
		if !g.canOccupy(cand, e) {
			continue
		}
		vx, vy := dirVector(d)
		align := fx*vx + fy*vy
		ncx := sx + tankSize/2
		ncy := sy + tankSize/2
		progress := (math.Abs(cx-targetX) + math.Abs(cy-targetY)) - (math.Abs(ncx-targetX) + math.Abs(ncy-targetY))
		crowdPenalty := g.crowdPenaltyAt(e, ncx, ncy)
		lineGap := math.Min(math.Abs(ncx-baseX), math.Abs(ncy-baseY))
		lineupBonus := clampF((38-lineGap)/38, 0, 1) * linePressure * 1.6
		score := align*2.5 + progress*0.08 + lineupBonus - crowdPenalty
		if d == e.dir {
			score += 0.15
		}
		if d == oppositeDir(e.dir) {
			score -= 0.35
		}
		if score > bestScore {
			bestScore = score
			best = d
		}
		if d == e.dir {
			currScore = score
		}
	}
	if best != e.dir && currScore > -math.MaxFloat64/2 {
		margin := 0.55
		if best == oppositeDir(e.dir) {
			margin = 1.25
		}
		if bestScore-currScore < margin {
			return e.dir
		}
	}
	return best
}

func (g *game) enemyTargetPoint(e *tank, _, cy float64) (float64, float64) {
	baseX := g.fort.box.x + g.fort.box.w/2
	baseY := g.fort.box.y + g.fort.box.h/2
	targetX := baseX
	targetY := baseY
	pressure := g.combinedFortressPressure(e)
	offset := (55 + 95*e.aiRand) * (0.24 + 0.76*(1-pressure))
	jitter := math.Sin(float64(e.age)*0.045+e.aiRand*7.3) * 12 * (1 - pressure*0.6)

	switch e.role {
	case roleLeftFlank:
		targetX = baseX - offset + jitter
	case roleRightFlank:
		targetX = baseX + offset + jitter
	}
	if cy > baseY-180 {
		targetX = baseX
	}
	return targetX, targetY
}

func (g *game) crowdPenaltyAt(self *tank, cx, cy float64) float64 {
	penalty := 0.0
	for _, other := range g.enemies {
		if other == self {
			continue
		}
		ox := other.x + tankSize/2
		oy := other.y + tankSize/2
		dist := math.Hypot(cx-ox, cy-oy)
		if dist < 96 {
			penalty += (96 - dist) / 96 * (1.6 + (1-self.aiRand)*0.9)
		}
	}
	return penalty
}

func (g *game) combinedFortressPressure(e *tank) float64 {
	p := enemyFortressPressure(e)*0.65 + playerSilencePressure(g.playerSilentFrames)*0.75
	return clampF(p, 0, 1)
}

func enemyFortressPressure(e *tank) float64 {
	rate := float64(e.age) / 1200.0
	if e.aggro > 0 {
		rate *= e.aggro
	}
	return clampF(rate, 0, 1)
}

func playerSilencePressure(frames int) float64 {
	return clampF(float64(frames)/1800.0, 0, 1)
}

func (g *game) enemyFortressHitProbability(e *tank, alignedBase bool) int {
	prob := 4
	prob += int(enemyFortressPressure(e) * 16)
	prob += int(playerSilencePressure(g.playerSilentFrames) * (22 + 10*e.aggro))
	if alignedBase {
		prob += 24
	} else if playerSilencePressure(g.playerSilentFrames) > 0.65 {
		prob += 6
	}
	if g.difficulty == diffHard {
		prob += 4
	}
	if prob < 3 {
		return 3
	}
	if prob > 92 {
		return 92
	}
	return prob
}

func (g *game) directionTowardFortress(e *tank) direction {
	ex := e.x + tankSize/2
	ey := e.y + tankSize/2
	fx := g.fort.box.x + g.fort.box.w/2
	fy := g.fort.box.y + g.fort.box.h/2
	dx := fx - ex
	dy := fy - ey
	if math.Abs(dx) > math.Abs(dy) {
		if dx < 0 {
			return left
		}
		return right
	}
	if dy < 0 {
		return up
	}
	return down
}

func (g *game) tryMoveTank(t *tank, dx, dy float64) bool {
	if dx == 0 && dy == 0 {
		return false
	}
	nx := t.x + dx
	ny := t.y + dy
	next := rect{nx, ny, tankSize, tankSize}
	if g.canOccupy(next, t) {
		t.x = nx
		t.y = ny
		return true
	}

	if dx != 0 && dy != 0 {
		nextX := rect{t.x + dx, t.y, tankSize, tankSize}
		if g.canOccupy(nextX, t) {
			t.x += dx
			return true
		}
		nextY := rect{t.x, t.y + dy, tankSize, tankSize}
		if g.canOccupy(nextY, t) {
			t.y += dy
			return true
		}
	}
	return false
}

func (g *game) canOccupy(next rect, self *tank) bool {
	if next.x < 0 || next.y < 0 || next.x+tankSize > screenW || next.y+tankSize > screenH {
		return false
	}
	for _, w := range g.walls {
		if overlap(next, w.box) {
			return false
		}
	}
	if overlap(next, g.fort.box) {
		return false
	}
	if self != &g.player && overlap(next, tankRect(g.player)) {
		return false
	}
	for _, e := range g.enemies {
		if e == self {
			continue
		}
		if overlap(next, tankRect(*e)) {
			return false
		}
	}
	return true
}

func (g *game) placeEnemy(e *tank, anchorX, anchorY float64) bool {
	offsets := []struct{ x, y float64 }{
		{0, 0}, {44, 0}, {-44, 0}, {0, 44}, {0, -44}, {62, 0}, {-62, 0}, {32, 32}, {-32, 32},
		{32, -32}, {-32, -32}, {86, 12}, {-86, 12},
	}
	for _, off := range offsets {
		candidate := rect{anchorX + off.x, anchorY + off.y, tankSize, tankSize}
		if g.canOccupy(candidate, e) {
			e.x = candidate.x
			e.y = candidate.y
			return true
		}
	}
	return false
}

func (g *game) enemyCrowdedAt(self *tank, cx, cy, dist float64) bool {
	for _, other := range g.enemies {
		if other == self {
			continue
		}
		ox := other.x + tankSize/2
		oy := other.y + tankSize/2
		if math.Abs(cx-ox)+math.Abs(cy-oy) < dist {
			return true
		}
	}
	return false
}

func nextStep(x, y float64, dir direction, speed float64) (float64, float64) {
	switch dir {
	case up:
		return x, y - speed
	case down:
		return x, y + speed
	case left:
		return x - speed, y
	default:
		return x + speed, y
	}
}

func dirVector(dir direction) (float64, float64) {
	switch dir {
	case up:
		return 0, -1
	case down:
		return 0, 1
	case left:
		return -1, 0
	default:
		return 1, 0
	}
}

func oppositeDir(dir direction) direction {
	switch dir {
	case up:
		return down
	case down:
		return up
	case left:
		return right
	default:
		return left
	}
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
