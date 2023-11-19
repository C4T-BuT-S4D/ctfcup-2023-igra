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

	return result
}

func (e *Engine) PushVectorX(from *geometry.Rectangle, object *geometry.Rectangle) *geometry.Vector {
	if !from.Intersects(object) {
		return &geometry.Vector{}
	}

	var vecs []*geometry.Vector

	if object.RightX > from.LeftX {
		vecs = append(vecs, &geometry.Vector{X: from.LeftX - object.RightX, Y: 0})
	}

	if object.LeftX < from.RightX {
		vecs = append(vecs, &geometry.Vector{X: from.RightX - object.LeftX, Y: 0})
	}

	v := vecs[0]
	for _, nv := range vecs {
		if nv.LengthSquared() < v.LengthSquared() {
			v = nv
		}
	}

	return v
}

func (e *Engine) PushVectorY(from *geometry.Rectangle, object *geometry.Rectangle) *geometry.Vector {
	if !from.Intersects(object) {
		return &geometry.Vector{}
	}

	var vecs []*geometry.Vector

	if object.BottomY > from.TopY {
		vecs = append(vecs, &geometry.Vector{X: 0, Y: from.TopY - object.BottomY})
	}

	if object.TopY < from.BottomY {
		vecs = append(vecs, &geometry.Vector{X: 0, Y: from.BottomY - object.TopY})
	}

	v := vecs[0]
	for _, nv := range vecs {
		if nv.LengthSquared() < v.LengthSquared() {
			v = nv
		}
	}

	return v
}
