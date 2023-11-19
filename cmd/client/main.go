package main

import (
	"fmt"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/engine"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/input"
)

// func (p *Player) Draw(screen *ebiten.Image) {
//	op := &ebiten.DrawImageOptions{}
//	op.GeoM.Translate(
//		float64(p.pos.X-p.img.Bounds().Dx()/2),
//		float64(p.pos.Y-p.img.Bounds().Dy()/2),
//	)
//	screen.DrawImage(p.img, op)
// }
//
// func (p *Player) Update() error {
//
//	if ebiten.IsKeyPressed(ebiten.KeyW) {
//		p.pos.Y -= 1
//	}
//	if ebiten.IsKeyPressed(ebiten.KeyA) {
//		p.pos.X -= 1
//	}
//	if ebiten.IsKeyPressed(ebiten.KeyS) {
//		p.pos.Y += 1
//	}
//	if ebiten.IsKeyPressed(ebiten.KeyD) {
//		p.pos.X += 1
//	}
//
//	return nil
// }

type Game struct {
	Engine *engine.Engine
}

func (g *Game) Update() error {
	var i input.Input

	if ebiten.IsKeyPressed(ebiten.KeyW) {
		i.WPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		i.APressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		i.SPressed = true
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		i.DPressed = true
	}

	if err := g.Engine.Update(&i); err != nil {
		return fmt.Errorf("updating engine state: %w", err)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.Engine.Draw(screen)
}

func (g *Game) Layout(_, _ int) (screenWidth, screenHeight int) {
	return 640, 480
}

func main() {
	e, err := engine.New()
	if err != nil {
		log.Fatal("initializing engine: %w", err)
	}

	g := Game{
		Engine: e,
	}

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(&g); err != nil {
		log.Fatalf("Failed to run game: %v", err)
	}
}
