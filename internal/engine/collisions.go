package engine

import (
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
)

func (e *Engine) Collisions(r *geometry.Rectangle) []object.GenericObject {
	var result []object.GenericObject

	if e.Player.Rectangle().Intersects(r) {
		result = append(result, e.Player)
	}

	for _, t := range e.Tiles {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	for _, t := range e.Items {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	for _, t := range e.Portals {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	for _, t := range e.Spikes {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	for _, t := range e.InvWalls {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	for _, t := range e.NPCs {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	if e.BossV1 != nil && e.BossV1.Rectangle().Intersects(r) {
		result = append(result, e.BossV1)
	}

	if e.BossV2 != nil && e.BossV2.Rectangle().Intersects(r) {
		result = append(result, e.BossV2)
	}

	for _, t := range e.EnemyBullets {
		if t.Rectangle().Intersects(r) {
			result = append(result, t)
		}
	}

	return result
}
