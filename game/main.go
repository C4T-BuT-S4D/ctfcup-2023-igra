package main

import (
	"embed"
	"github.com/Rulox/ebitmx"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/salviati/go-tmx/tmx"
	"image"
	"image/color"
	"log"

	_ "image/png"
)

//go:embed tiles
var embeddedFS embed.FS

//func (p *Player) Draw(screen *ebiten.Image) {
//	op := &ebiten.DrawImageOptions{}
//	op.GeoM.Translate(
//		float64(p.pos.X-p.img.Bounds().Dx()/2),
//		float64(p.pos.Y-p.img.Bounds().Dy()/2),
//	)
//	screen.DrawImage(p.img, op)
//}
//
//func (p *Player) Update() error {
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
//}

type Game struct {
	gameMap     *tmx.Map
	tileSet     *ebitmx.EbitenTileset
	resultImage *ebiten.Image
	player      *Player
	world       *World
}

func (g *Game) Update() error {
	currentPos := g.world.Player.Rectangle()
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		currentPos.TopY -= 1
		currentPos.BottomY -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) {
		currentPos.LeftX -= 1
		currentPos.RightX -= 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) {
		currentPos.TopY += 1
		currentPos.BottomY += 1
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) {
		currentPos.LeftX += 1
		currentPos.RightX += 1
	}

	intersects := g.world.Intersects(currentPos)
	// TODO: Handle collision for each object type.
	if len(intersects) > 0 {
		return nil
	}

	g.world.Player.SetRectangle(currentPos)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, o := range g.world.Objects {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			float64(o.Rectangle().LeftX),
			float64(o.Rectangle().TopY),
		)
		screen.DrawImage(o.Image(), op)
	}
	//for _, l := range g.gameMap.Layers {
	//	if l.Name != "tiles" {
	//		continue
	//	}
	//
	//	for x := 0; x < g.gameMap.Width; x++ {
	//		for y := 0; y < g.gameMap.Height; y++ {
	//			dt := l.DecodedTiles[y*g.gameMap.Width+x]
	//			//log.Printf("x = %d, y = %d, isNil: %v, tile: %d\n", x, y, dt.IsNil(), dt.ID)
	//			if dt.IsNil() {
	//				continue
	//			}
	//			tileID := dt.ID
	//			op := &ebiten.DrawImageOptions{}
	//			op.GeoM.Translate(
	//				float64(x*g.gameMap.TileWidth),
	//				float64(y*g.gameMap.TileHeight),
	//			)
	//			screen.DrawImage(g.getTileImgByID(tileID), op)
	//		}
	//	}
	//}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(
		float64(g.world.Player.Rectangle().LeftX),
		float64(g.world.Player.Rectangle().TopY),
	)
	screen.DrawImage(g.world.Player.Image(), op)

	//cx, cy := ebiten.CursorPosition()
	//ebitenutil.DebugPrint(
	//	screen,
	//	fmt.Sprintf("cx:%d, cy:%d\ntype: %s\n", cx, cy, g.currentTileType),
	//)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}

func getTileImgByID(_id tmx.ID, tileSet *ebitmx.EbitenTileset, img *ebiten.Image) *ebiten.Image {
	// The tsx format starts counting tiles from 1, so to make these calculations
	// work correctly, we need to decrement the ID by 1
	id := int(_id)
	//id -= 1

	x0 := (id % tileSet.TilesetWidth) * tileSet.TileWidth
	y0 := (id / tileSet.TilesetWidth) * tileSet.TileHeight
	x1, y1 := x0+tileSet.TileWidth, y0+tileSet.TileHeight

	return img.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image)
}

func findPlayerSpawn(map_ *tmx.Map) Point {
	for _, og := range map_.ObjectGroups {
		for _, o := range og.Objects {
			for _, p := range o.Properties {
				if p.Name == "type" && p.Value == "player_spawn" {
					return Point{int(o.X), int(o.Y)}
				}
			}
		}
	}

	panic("No player spawn found")
}

func main() {
	var resultImage *ebiten.Image
	{
		imgFile, err := embeddedFS.Open("tiles/result.png")
		if err != nil {
			log.Fatalf("Failed to open result.png: %v", err)
		}

		img, _, err := image.Decode(imgFile)
		if err != nil {
			log.Fatalf("Failed to decode image: %v", err)
		}

		resultImage = ebiten.NewImageFromImage(img)
	}

	mapFile, err := embeddedFS.Open("tiles/test.tmx")
	if err != nil {
		log.Fatalf("Failed to open map file: %v", err)
	}

	testMap, err := tmx.Read(mapFile)
	if err != nil {
		log.Fatalf("Failed to load map: %v", err)
	}

	iceTiles, err := ebitmx.GetTilesetFromFS(embeddedFS, "tiles/ice.tsx")
	if err != nil {
		log.Fatalf("Failed to load tiles: %v", err)
	}

	pimg := ebiten.NewImage(16, 16)
	pimg.Fill(color.White)

	world := World{}
	for _, l := range testMap.Layers {
		for x := 0; x < testMap.Width; x++ {
			for y := 0; y < testMap.Height; y++ {
				dt := l.DecodedTiles[y*testMap.Width+x]
				if dt.IsNil() {
					continue
				}

				tileID := dt.ID
				world.AddObject(StaticTile{
					width:  testMap.TileWidth,
					height: testMap.TileHeight,
					x:      x * testMap.TileWidth,
					y:      y * testMap.TileHeight,
					img:    getTileImgByID(tileID, iceTiles, resultImage),
				})
			}
		}
	}

	ppos := findPlayerSpawn(testMap)
	world.Player = &Player{
		img:    pimg,
		width:  pimg.Bounds().Dx(),
		height: pimg.Bounds().Dy(),
		x:      float64(ppos.X),
		y:      float64(ppos.Y),
	}

	g := Game{
		gameMap:     testMap,
		tileSet:     iceTiles,
		resultImage: resultImage,
		world:       &world,
	}
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(&g); err != nil {
		log.Fatalf("Failed to run game: %v", err)
	}
}
