package engine

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"

	"github.com/Rulox/ebitmx"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/salviati/go-tmx/tmx"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/input"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/item"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/physics"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/player"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/resources"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/tiles"

	// Register png codec.
	_ "image/png"
)

type Factory func() (*Engine, error)

type Engine struct {
	Tiles   []*tiles.StaticTile `json:"tiles"`
	Player  *player.Player      `json:"player"`
	Items   []*item.Item        `json:"items"`
	signKey []byte              `json:"-"`
}

func getProperties(o *tmx.Object) map[string]string {
	properties := make(map[string]string)
	for _, p := range o.Properties {
		properties[p.Name] = p.Value
	}
	return properties
}

var ErrNoPlayerSpawn = errors.New("no player spawn found")

func findPlayerSpawn(tileMap *tmx.Map) (*geometry.Point, error) {
	for _, og := range tileMap.ObjectGroups {
		for _, o := range og.Objects {
			if getProperties(&o)["type"] == "player_spawn" {
				return &geometry.Point{
					X: o.X,
					Y: o.Y,
				}, nil
			}
		}
	}

	return nil, ErrNoPlayerSpawn
}

func getTileImgByID(tileID tmx.ID, tileSet *ebitmx.EbitenTileset, img *ebiten.Image) *ebiten.Image {
	id := int(tileID)

	x0 := (id % tileSet.TilesetWidth) * tileSet.TileWidth
	y0 := (id / tileSet.TilesetWidth) * tileSet.TileHeight
	x1, y1 := x0+tileSet.TileWidth, y0+tileSet.TileHeight

	return img.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image)
}

func New(signKey []byte) (*Engine, error) {
	var resultImage *ebiten.Image
	imgFile, err := resources.EmbeddedFS.Open("tiles/result.png")
	if err != nil {
		return nil, fmt.Errorf("failed to open tileset: %w", err)
	}

	img, _, err := image.Decode(imgFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode tileset: %w", err)
	}

	resultImage = ebiten.NewImageFromImage(img)

	mapFile, err := resources.EmbeddedFS.Open("tiles/test.tmx")
	if err != nil {
		return nil, fmt.Errorf("failed to open map: %w", err)
	}

	testMap, err := tmx.Read(mapFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode map: %w", err)
	}

	iceTiles, err := ebitmx.GetTilesetFromFS(resources.EmbeddedFS, "tiles/ice.tsx")
	if err != nil {
		return nil, fmt.Errorf("failed to get tileset: %w", err)
	}

	var mapTiles []*tiles.StaticTile

	for _, l := range testMap.Layers {
		for x := 0; x < testMap.Width; x++ {
			for y := 0; y < testMap.Height; y++ {
				dt := l.DecodedTiles[y*testMap.Width+x]
				if dt.IsNil() {
					continue
				}

				mapTiles = append(
					mapTiles,
					tiles.NewStaticTile(
						&geometry.Point{
							X: float64(x * testMap.TileWidth),
							Y: float64(y * testMap.TileHeight),
						},
						testMap.TileWidth,
						testMap.TileHeight,
						getTileImgByID(dt.ID, iceTiles, resultImage),
					),
				)
			}
		}
	}

	playerPos, err := findPlayerSpawn(testMap)
	if err != nil {
		return nil, fmt.Errorf("can't find player position: %w", err)
	}

	p := player.New(playerPos)

	var items []*item.Item

	for _, og := range testMap.ObjectGroups {
		for _, o := range og.Objects {
			props := getProperties(&o)
			if props["type"] == "item" {
				items = append(items, item.New(&geometry.Point{
					X: o.X,
					Y: o.Y,
				}, int(o.Width), int(o.Height), props["name"], false))
			}
		}
	}

	return &Engine{
		Tiles:  mapTiles,
		Player: p,
		Items:  items,
	}, nil
}

var ErrIncorrectHmac = errors.New("incorrect snapshot hmac")

func NewFromSnapshot(signKey []byte, snapshot *Snapshot) (*Engine, error) {
	mac := hmac.New(sha256.New, signKey)
	if _, err := mac.Write(snapshot.Data); err != nil {
		return nil, fmt.Errorf("signing data: %w", err)
	}

	if !hmac.Equal(mac.Sum(nil), snapshot.Sign) {
		return nil, ErrIncorrectHmac
	}

	return nil, nil
}

type Snapshot struct {
	Data []byte
	Sign []byte
}

func (e *Engine) MakeSnapshot() (*Snapshot, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("marshalling engine: %w", err)
	}

	mac := hmac.New(sha256.New, e.signKey)
	if _, err := mac.Write(data); err != nil {
		return nil, fmt.Errorf("signing data: %w", err)
	}

	return &Snapshot{
		Data: data,
		Sign: mac.Sum(nil),
	}, nil
}

func (s *Snapshot) Resign(signKey []byte) error {
	mac := hmac.New(sha256.New, signKey)
	if _, err := mac.Write(s.Data); err != nil {
		return fmt.Errorf("signing data: %w", err)
	}

	s.Sign = mac.Sum(nil)
	return nil
}

func (e *Engine) Draw(screen *ebiten.Image) {
	for _, t := range e.Tiles {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			t.Rectangle().LeftX,
			t.Rectangle().TopY,
		)
		screen.DrawImage(t.Image, op)
	}

	for _, it := range e.Items {
		if it.Collected {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			it.Rectangle().LeftX,
			it.Rectangle().TopY,
		)
		screen.DrawImage(it.Image, op)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(
		e.Player.Rectangle().LeftX,
		e.Player.Rectangle().TopY,
	)
	screen.DrawImage(e.Player.Image, op)
}

func (e *Engine) Update(inp *input.Input) error {
	e.ProcessPlayerInput(inp)
	e.Player.Move(&geometry.Vector{X: e.Player.Speed.X, Y: 0})
	e.AlignPlayerX()
	e.Player.Move(&geometry.Vector{X: 0, Y: e.Player.Speed.Y})
	e.AlignPlayerY()
	e.CollectItems()

	return nil
}

func (e *Engine) ProcessPlayerInput(inp *input.Input) {
	if e.Player.OnGround {
		e.Player.Acceleration.Y = 0
	} else {
		e.Player.Acceleration.Y = physics.GravityAcceleration
	}

	if (inp.IsKeyPressed(ebiten.KeySpace) || inp.IsKeyPressed(ebiten.KeyW)) && e.Player.OnGround {
		e.Player.Speed.Y = -5
	}

	switch {
	case inp.IsKeyPressed(ebiten.KeyA):
		e.Player.Speed.X = -2.5
	case inp.IsKeyPressed(ebiten.KeyD):
		e.Player.Speed.X = 2.5
	default:
		e.Player.Speed.X = 0
	}

	e.Player.ApplyAcceleration()
}

func (e *Engine) AlignPlayerX() {
	var pv *geometry.Vector

	for _, c := range e.Collisions(e.Player.Rectangle()) {
		if c.Type() != object.StaticTileType {
			continue
		}

		pv = c.Rectangle().PushVectorX(e.Player.Rectangle())
		break
	}

	if pv == nil {
		return
	}

	e.Player.Move(pv)
}

func (e *Engine) AlignPlayerY() {
	var pv *geometry.Vector

	for _, c := range e.Collisions(e.Player.Rectangle()) {
		if c.Type() != object.StaticTileType {
			continue
		}

		pv = c.Rectangle().PushVectorY(e.Player.Rectangle())
		break
	}

	e.Player.OnGround = false

	if pv == nil {
		return
	}

	e.Player.Move(pv)

	if pv.Y < 0 {
		e.Player.OnGround = true
	} else {
		e.Player.Speed.Y = 0
	}
}

func (e *Engine) CollectItems() {
	for _, c := range e.Collisions(e.Player.Rectangle()) {
		if c.Type() != object.Item {
			continue
		}

		it := c.(*item.Item)
		if it.Collected {
			continue
		}

		e.Player.Collect(it)
	}
}

func (e *Engine) Checksum() (string, error) {
	if snapshot, err := e.MakeSnapshot(); err != nil {
		return "", fmt.Errorf("making snapshot: %w", err)
	} else {
		hash := sha256.New()
		if _, err := hash.Write(snapshot.Data); err != nil {
			return "", fmt.Errorf("hashing snapshot: %w", err)
		}

		return hex.EncodeToString(hash.Sum(snapshot.Data)), nil
	}
}

var ErrInvalidChecksum = errors.New("invalid checksum")

func (e *Engine) ValidateChecksum(checksum string) error {
	if currentChecksum, err := e.Checksum(); err != nil {
		return fmt.Errorf("getting correct checksum: %w", err)
	} else if currentChecksum != checksum {
		return ErrInvalidChecksum
	}

	return nil
}
