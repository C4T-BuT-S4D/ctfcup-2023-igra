package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"time"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/camera"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/damage"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/portal"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/sprites"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/wall"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/samber/lo"

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
	gameserverpb "github.com/c4t-but-s4d/ctfcup-2023-igra/proto/go/gameserver"

	// Register png codec.
	_ "image/png"
)

type Factory func() (*Engine, error)

type Config struct {
	SnapshotsDir string
	Level        string
}

type Engine struct {
	Tiles    []*tiles.StaticTile `json:"-"`
	Camera   *camera.Camera      `json:"-"`
	Player   *player.Player      `json:"player"`
	Items    []*item.Item        `json:"items"`
	Portals  []*portal.Portal    `json:"-"`
	Spikes   []*damage.Spike     `json:"-"`
	InvWalls []*wall.InvWall     `json:"-"`

	StartSnapshot *Snapshot `json:"-"`
	snapshotsDir  string
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
			if o.Type == "player_spawn" {
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

func New(config Config, spriteManager *sprites.Manager) (*Engine, error) {
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

	mapFile, err := resources.EmbeddedFS.Open(fmt.Sprintf("tiles/%s.tmx", config.Level))
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

	playerSprite, err := spriteManager.GetSprite(sprites.Player)
	if err != nil {
		return nil, fmt.Errorf("getting player sprite: %w", err)
	}

	p := player.New(playerPos, playerSprite)

	var items []*item.Item
	var spikes []*damage.Spike
	var invwalls []*wall.InvWall
	portalsMap := make(map[string]*portal.Portal)

	for _, og := range testMap.ObjectGroups {
		for _, o := range og.Objects {
			props := getProperties(&o)
			switch o.Type {
			case "item":
				items = append(items, item.New(
					&geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					o.Width,
					o.Height,
					o.Name,
					false,
				))
			case "portal":
				portalsMap[o.Name] = portal.New(
					&geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					o.Width,
					o.Height,
					props["portal-to"],
					nil)
			case "spike":
				img, err := spriteManager.GetSprite(sprites.Spike)
				if err != nil {
					return nil, fmt.Errorf("getting spike sprite: %w", err)
				}

				spikes = append(spikes, damage.NewSpike(
					&geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					img,
					o.Width,
					o.Height,
				))
			case "invwall":
				invwalls = append(invwalls, wall.NewInvWall(&geometry.Point{
					X: o.X,
					Y: o.Y,
				},
					o.Width,
					o.Height))
			}
		}
	}
	for name, p := range portalsMap {
		if p.PortalTo == "" {
			continue
		}
		toPortal := portalsMap[p.PortalTo]
		if toPortal == nil {
			return nil, fmt.Errorf("destination %s not found for portal %s", p.PortalTo, name)
		}
		p.TeleportTo = toPortal.Origin.Add(&geometry.Vector{
			X: 32,
			Y: 0,
		})
	}

	cam := &camera.Camera{
		Object: &object.Object{
			Origin: &geometry.Point{
				X: 0,
				Y: 0,
			},
			Width:  camera.WIDTH,
			Height: camera.HEIGHT,
		},
	}

	return &Engine{
		Tiles:        mapTiles,
		Camera:       cam,
		Player:       p,
		Items:        items,
		Portals:      lo.Values(portalsMap),
		Spikes:       spikes,
		InvWalls:     invwalls,
		snapshotsDir: config.SnapshotsDir,
	}, nil
}

func NewFromSnapshot(config Config, snapshot *Snapshot, spritesManager *sprites.Manager) (*Engine, error) {
	e, err := New(config, spritesManager)
	if err != nil {
		return nil, fmt.Errorf("creating engine: %w", err)
	}

	e.StartSnapshot = snapshot

	if err := json.Unmarshal(snapshot.Data, e); err != nil {
		return nil, fmt.Errorf("applying snapshot: %w", err)
	}

	return e, nil
}

type Snapshot struct {
	Data []byte
}

func NewSnapshotFromProto(proto *gameserverpb.EngineSnapshot) *Snapshot {
	return &Snapshot{Data: proto.Data}
}

func (s *Snapshot) ToProto() *gameserverpb.EngineSnapshot {
	if s == nil {
		return nil
	}
	return &gameserverpb.EngineSnapshot{
		Data: s.Data,
	}
}

func (e *Engine) MakeSnapshot() (*Snapshot, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("marshalling engine: %w", err)
	}

	return &Snapshot{
		Data: data,
	}, nil
}

func (e *Engine) SaveSnapshot(snapshot *Snapshot) error {
	if e.snapshotsDir == "" {
		return nil
	}

	filename := fmt.Sprintf("snapshot_%s", time.Now().UTC().Format("2006-01-02T15:04:05.999999999"))

	if err := os.WriteFile(filepath.Join(e.snapshotsDir, filename), snapshot.Data, 0o400); err != nil {
		return fmt.Errorf("writing snapshot file: %w", err)
	}

	return nil
}

func (e *Engine) Draw(screen *ebiten.Image) {
	if e.Player.IsDead() {
		// Draw "YOU DIED" text over all screen using red color.
		img := ebiten.NewImageFromImage(screen)
		img.Fill(color.RGBA{0x80, 0x80, 0x80, 0xff})
		text := "YOU DIED"

		ebitenutil.DebugPrintAt(img, text, 0, 0)
		screen.DrawImage(img, nil)
		return
	}

	for _, c := range e.Collisions(e.Camera.Rectangle()) {
		visible := c.Rectangle().Sub(e.Camera.Rectangle())
		base := geometry.Origin.Add(visible)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			base.X,
			base.Y,
		)

		switch c.Type() {
		case object.StaticTileType:
			t := c.(*tiles.StaticTile)
			screen.DrawImage(t.Image, op)
		case object.Item:
			it := c.(*item.Item)
			if it.Collected {
				continue
			}
			screen.DrawImage(it.Image, op)
		case object.PlayerType:
			screen.DrawImage(e.Player.Image, op)
		case object.Portal:
			p := c.(*portal.Portal)
			screen.DrawImage(p.Image, op)
		case object.Spike:
			d := c.(*damage.Spike)
			screen.DrawImage(d.Image, op)
		}
	}
}

func (e *Engine) Update(inp *input.Input) error {
	if e.Player.IsDead() {
		return nil
	}

	e.ProcessPlayerInput(inp)
	e.Player.Move(&geometry.Vector{X: e.Player.Speed.X, Y: 0})
	e.AlignPlayerX()
	e.Player.Move(&geometry.Vector{X: 0, Y: e.Player.Speed.Y})
	e.AlignPlayerY()
	e.CheckPortals()
	e.CheckSpikes()
	if err := e.CollectItems(); err != nil {
		return fmt.Errorf("collecting items: %w", err)
	}
	e.Camera.MoveTo(e.Player.Origin.Add(&geometry.Vector{
		X: -camera.WIDTH/2 + e.Player.Width/2,
		Y: -camera.HEIGHT/2 + e.Player.Height/2,
	}))

	return nil
}

func (e *Engine) ProcessPlayerInput(inp *input.Input) {
	if e.Player.OnGround {
		e.Player.Acceleration.Y = 0
	} else {
		e.Player.Acceleration.Y = physics.GravityAcceleration
	}

	if (inp.IsKeyPressed(ebiten.KeySpace) || inp.IsKeyPressed(ebiten.KeyW)) && e.Player.OnGround {
		e.Player.Speed.Y = -5 * 2
	}

	switch {
	case inp.IsKeyPressed(ebiten.KeyA):
		e.Player.Speed.X = -2.5 * 2
	case inp.IsKeyPressed(ebiten.KeyD):
		e.Player.Speed.X = 2.5 * 2
	default:
		e.Player.Speed.X = 0
	}

	e.Player.ApplyAcceleration()
}

func (e *Engine) AlignPlayerX() {
	var pv *geometry.Vector

	for _, c := range e.Collisions(e.Player.Rectangle()) {
		if c.Type() != object.StaticTileType && c.Type() != object.InvWall {
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
		if c.Type() != object.StaticTileType && c.Type() != object.InvWall {
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

func (e *Engine) CollectItems() error {
	collectedSomething := false

	for _, c := range e.Collisions(e.Player.Rectangle()) {
		if c.Type() != object.Item {
			continue
		}

		it := c.(*item.Item)
		if it.Collected {
			continue
		}

		e.Player.Collect(it)

		collectedSomething = true
	}

	if collectedSomething {
		snapshot, err := e.MakeSnapshot()
		if err != nil {
			return fmt.Errorf("making snapshot: %w", err)
		}

		if err := e.SaveSnapshot(snapshot); err != nil {
			return fmt.Errorf("saving snapshot: %w", err)
		}
	}

	return nil
}

func (e *Engine) CheckPortals() {
	for _, c := range e.Collisions(e.Player.Rectangle()) {
		if c.Type() != object.Portal {
			continue
		}

		p := c.(*portal.Portal)
		if p.TeleportTo == nil {
			continue
		}
		e.Player.MoveTo(p.TeleportTo)
	}
}

func (e *Engine) CheckSpikes() {
	for _, c := range e.Collisions(e.Player.Rectangle()) {
		if c.Type() != object.Spike {
			continue
		}

		s := c.(*damage.Spike)
		e.Player.Health -= s.Damage
	}
}

func (e *Engine) Checksum() (string, error) {
	snapshot, err := e.MakeSnapshot()
	if err != nil {
		return "", fmt.Errorf("making snapshot: %w", err)
	}

	hash := sha256.New()
	if _, err := hash.Write(snapshot.Data); err != nil {
		return "", fmt.Errorf("hashing snapshot: %w", err)
	}

	return hex.EncodeToString(hash.Sum(snapshot.Data)), nil
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
