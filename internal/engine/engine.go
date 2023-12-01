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
	"strings"
	"time"

	"github.com/Rulox/ebitmx"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/salviati/go-tmx/tmx"
	"github.com/samber/lo"
	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/image/font"
	"slices"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/camera"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/damage"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/dialog"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/fonts"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/input"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/item"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/npc"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/physics"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/player"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/portal"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/resources"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/sprites"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/tiles"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/wall"
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
	Tiles    []*tiles.StaticTile `json:"-" msgpack:"-"`
	Camera   *camera.Camera      `json:"-" msgpack:"camera"`
	Player   *player.Player      `json:"player" msgpack:"player"`
	Items    []*item.Item        `json:"items" msgpack:"items"`
	Portals  []*portal.Portal    `json:"-" msgpack:"portals"`
	Spikes   []*damage.Spike     `json:"-" msgpack:"spikes"`
	InvWalls []*wall.InvWall     `json:"-" msgpack:"invWalls"`
	NPCs     []*npc.NPC          `json:"-" msgpack:"npcs"`

	StartSnapshot *Snapshot `json:"-" msgpack:"-"`

	fontsManager      *fonts.Manager
	snapshotsDir      string
	playerSpawn       *geometry.Point
	activeNPC         *npc.NPC
	dialogInputBuffer []string

	Paused bool `msgpack:"paused"`
	Tick   int  `msgpack:"tick"`
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

func New(config Config, spriteManager *sprites.Manager, fontsManager *fonts.Manager) (*Engine, error) {
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

	mapFile, err := resources.EmbeddedFS.Open(fmt.Sprintf("levels/%s.tmx", config.Level))
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

	p, err := player.New(playerPos, spriteManager)
	if err != nil {
		return nil, fmt.Errorf("creating player: %w", err)
	}

	var items []*item.Item
	var spikes []*damage.Spike
	var invwalls []*wall.InvWall
	var npcs []*npc.NPC
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
				img, err := spriteManager.GetSprite(sprites.Portal)
				if err != nil {
					return nil, fmt.Errorf("getting portal sprite: %w", err)
				}
				portalsMap[o.Name] = portal.New(
					&geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					img,
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
			case "npc":
				img, err := spriteManager.GetSprite(sprites.Type(props["sprite"]))
				if err != nil {
					return nil, fmt.Errorf("getting slon sprite: %w", err)
				}
				dimg, err := spriteManager.GetSprite(sprites.Type(props["dialog-sprite"]))
				if err != nil {
					return nil, fmt.Errorf("getting slon dialog sprite: %w", err)
				}
				npcs = append(npcs, npc.New(
					&geometry.Point{
						X: o.X,
						Y: o.Y,
					},
					img,
					dimg,
					o.Width,
					o.Height,
					dialog.NewDummy("Hello, I'm SLONIK! pröööh об этом"),
				))
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

	keys := lo.Keys(portalsMap)
	slices.Sort(keys)
	portals := make([]*portal.Portal, 0, len(keys))
	for _, key := range keys {
		portals = append(portals, portalsMap[key])
	}

	return &Engine{
		Tiles:        mapTiles,
		Camera:       cam,
		Player:       p,
		Items:        items,
		Portals:      portals,
		Spikes:       spikes,
		InvWalls:     invwalls,
		NPCs:         npcs,
		fontsManager: fontsManager,
		snapshotsDir: config.SnapshotsDir,
		playerSpawn:  playerPos,
	}, nil
}

func NewFromSnapshot(config Config, snapshot *Snapshot, spritesManager *sprites.Manager, fontsManager *fonts.Manager) (*Engine, error) {
	e, err := New(config, spritesManager, fontsManager)
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

func (e *Engine) Reset() {
	e.Player.MoveTo(e.playerSpawn)
	e.Player.Health = player.DefaultHealth
	e.activeNPC = nil
	e.Tick = 0
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

func (e *Engine) drawDiedScreen(screen *ebiten.Image) {
	// Draw "YOU DIED" text over all screen using red color.
	face := e.fontsManager.Get(fonts.DSouls)
	redColor := color.RGBA{R: 255, G: 0, B: 0, A: 255}

	width := font.MeasureString(face, "YOU DIED")

	text.Draw(screen, "YOU DIED", face, camera.WIDTH/2-width.Floor()/2, camera.HEIGHT/2, redColor)
}

func (e *Engine) drawNPCDialog(screen *ebiten.Image) {
	colorWhite := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	// Draw dialog border (outer rectangle).
	borderw, borderh := camera.WIDTH-camera.WIDTH/8, camera.HEIGHT/4
	img := ebiten.NewImage(borderw, borderh)
	img.Fill(colorWhite)
	op := &ebiten.DrawImageOptions{}
	bx, by := camera.WIDTH/16.0, camera.HEIGHT/4.0*3-camera.HEIGHT/16
	op.GeoM.Translate(bx, by)
	screen.DrawImage(img, op)

	// Draw dialog border (inner rectangle).
	ibw, ibh := borderw-camera.WIDTH/32, borderh-camera.HEIGHT/32
	ibx, iby := bx+camera.WIDTH/64, by+camera.HEIGHT/64
	img = ebiten.NewImage(ibw, ibh)
	img.Fill(color.RGBA{R: 0, G: 0, B: 0, A: 0xff})
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(ibx, iby)
	screen.DrawImage(img, op)

	// Draw dialog NPC image.
	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(camera.WIDTH/2+camera.WIDTH/8, camera.HEIGHT/2)
	screen.DrawImage(e.activeNPC.DialogImage, op)

	// TODO(jnovikov): show only last X lines ?
	// Draw dialog text.
	dtx, dty := ibx+camera.WIDTH/32, iby+camera.HEIGHT/32
	face := e.fontsManager.Get(fonts.Dialog)
	txt := e.activeNPC.Dialog.State().Text
	text.Draw(screen, txt, face, int(dtx), int(dty), colorWhite)

	// Draw dialog input buffer.
	if len(e.dialogInputBuffer) > 0 {
		nLines := strings.Count(txt, "\n")
		dtbx, dtby := dtx, dty+float64(nLines*face.Metrics().Height.Floor())+1.0*float64(face.Metrics().Height.Floor())
		c := color.RGBA{R: 0x00, G: 0xff, B: 0xff, A: 0xff}
		text.Draw(screen, strings.Join(e.dialogInputBuffer, ""), face, int(dtbx), int(dtby), c)
	}
}

func (e *Engine) Draw(screen *ebiten.Image) {
	if e.Player.IsDead() {
		e.drawDiedScreen(screen)
		return
	}

	for _, c := range e.Collisions(e.Camera.Rectangle()) {
		visible := c.Rectangle().Sub(e.Camera.Rectangle())
		base := geometry.Origin.Add(visible)
		op := &ebiten.DrawImageOptions{}

		if c.Type() == object.PlayerType && e.Player.LooksRight {
			op.GeoM.Scale(-1, 1)
			op.GeoM.Translate(e.Player.Width, 0)
		}

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
			screen.DrawImage(e.Player.Image(), op)
		case object.Portal:
			p := c.(*portal.Portal)
			screen.DrawImage(p.Image, op)
		case object.Spike:
			d := c.(*damage.Spike)
			screen.DrawImage(d.Image, op)
		case object.NPC:
			n := c.(*npc.NPC)
			screen.DrawImage(n.Image, op)
		default:
			// not an item.
		}
	}

	if e.activeNPC != nil {
		e.drawNPCDialog(screen)
	}
}

func (e *Engine) Update(inp *input.Input) error {
	e.Tick++

	if e.Paused {
		if inp.IsKeyNewlyPressed(ebiten.KeyP) {
			e.Paused = false
		} else {
			return nil
		}
	} else if inp.IsKeyNewlyPressed(ebiten.KeyP) {
		e.Paused = true
		e.Player.Speed = &geometry.Vector{}
	}

	if inp.IsKeyNewlyPressed(ebiten.KeyR) {
		e.Reset()
		return nil
	}

	if e.Player.IsDead() {
		return nil
	}

	if e.activeNPC != nil {
		if inp.IsKeyNewlyPressed(ebiten.KeyEscape) {
			e.activeNPC = nil
			e.dialogInputBuffer = e.dialogInputBuffer[:0]
			return nil
		}
		pk := inp.JustPressedKeys()
		if len(pk) > 0 {
			c := pk[0]
			switch c {
			case ebiten.KeyBackspace:
				// backspace
				if len(e.dialogInputBuffer) > 0 {
					e.dialogInputBuffer = e.dialogInputBuffer[:len(e.dialogInputBuffer)-1]
				}
			case ebiten.KeyEnter:
				// enter
				e.activeNPC.Dialog.Feed(strings.Join(e.dialogInputBuffer, ""))
				e.dialogInputBuffer = e.dialogInputBuffer[:0]
			default:
				// TODO(jnovikov): rework this.
				e.dialogInputBuffer = append(e.dialogInputBuffer, c.String())
			}
		}

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

	availableNPC := e.CheckNPCClose()
	if availableNPC != nil && inp.IsKeyNewlyPressed(ebiten.KeyE) {
		e.activeNPC = availableNPC
		e.activeNPC.Dialog.Greeting()
		return nil
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
		e.Player.LooksRight = false
	case inp.IsKeyPressed(ebiten.KeyD):
		e.Player.Speed.X = 2.5 * 2
		e.Player.LooksRight = true
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

func (e *Engine) CheckNPCClose() *npc.NPC {
	for _, c := range e.Collisions(e.Player.Rectangle().Extended(40)) {
		if c.Type() != object.NPC {
			continue
		}

		n := c.(*npc.NPC)
		return n
	}

	return nil
}

func (e *Engine) Checksum() (string, error) {
	b, err := msgpack.Marshal(e)
	if err != nil {
		return "", fmt.Errorf("marshalling engine: %w", err)
	}

	hash := sha256.New()
	if _, err := hash.Write(b); err != nil {
		return "", fmt.Errorf("hashing snapshot: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
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
