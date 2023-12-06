package boss

import (
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/damage"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/geometry"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/item"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/object"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/player"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/portal"
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	BossV2Width         = 128
	BossV2Height        = 128
	BossV2CrossWidth    = 128
	BossV2CrossHeight   = 128
	BossV2BarbellWidth  = 128
	BossV2BarbellHeight = 128
)

type CrossInfo struct {
	Obj       *object.Object
	Tick      int
	Direction *geometry.Vector
	Ticks     int
}

type BarbellInfo struct {
	Obj        *object.Object
	FallTicks  int
	FallSpeed  float64
	State      int
	StartTicks int
	Ticks      int
}

type V2 struct {
	*object.Object `json:"-"`
	Name           string `json:"-"`
	bulletImage    *ebiten.Image
	Image          *ebiten.Image   `json:"-" msgpack:"-"`
	Ticks          int             `json:"-"`
	StartHealth    int64           `json:"-"`
	Health         int64           `json:"-"`
	Dead           bool            `json:"dead"`
	WinPoint       *geometry.Point `json:"-"`
	PortalName     string          `json:"-"`
	ItemName       string          `json:"-"`
	Portal         *portal.Portal  `json:"-"`
	Item           *item.Item      `json:"-"`

	StartPoint *geometry.Point `json:"-"`
	Speed      float64         `json:"-"`
	Width      float64         `json:"-" msgpack:"width_"`
	Height     float64         `json:"-" msgpack:"height_"`
	State      int             `json:"-"`
	Crosses    []*CrossInfo    `json:"-"`
	Barbells   []*BarbellInfo  `json:"-"`
}

func (v *V2) Type() object.Type {
	return object.BossV2
}

func NewV2(name string, origin *geometry.Point, img *ebiten.Image, bulletImage *ebiten.Image, speed float64, width float64, height float64, health int64, portalName string, itemName string) *V2 {
	v := &V2{
		Object: &object.Object{
			Origin: origin,
			Width:  BossV2Width,
			Height: BossV2Height,
		},
		Name:        name,
		bulletImage: bulletImage,
		Image:       img,
		StartHealth: health,
		Health:      health,
		PortalName:  portalName,
		ItemName:    itemName,

		StartPoint: origin,
		Speed:      speed,
		Width:      width,
		Height:     height,
		Crosses:    nil,
	}

	v.setCrosses()
	v.setBarbells()

	return v
}

func (v *V2) setCrosses() {
	v.Crosses = []*CrossInfo{
		{
			Obj: &object.Object{
				Origin: &geometry.Point{X: v.StartPoint.X, Y: v.StartPoint.Y},
				Width:  BossV2CrossWidth,
				Height: BossV2CrossHeight,
			},
			Ticks: 150,
		},
		{
			Obj: &object.Object{
				Origin: &geometry.Point{X: v.StartPoint.X - v.Width, Y: v.StartPoint.Y},
				Width:  BossV2CrossWidth,
				Height: BossV2CrossHeight,
			},
			Ticks: 170,
		},
		{
			Obj: &object.Object{
				Origin: &geometry.Point{X: v.StartPoint.X, Y: v.StartPoint.Y + v.Height},
				Width:  BossV2CrossWidth,
				Height: BossV2CrossHeight,
			},
			Ticks: 190,
		},
		{
			Obj: &object.Object{
				Origin: &geometry.Point{X: v.StartPoint.X - v.Width, Y: v.StartPoint.Y + v.Height},
				Width:  BossV2CrossWidth,
				Height: BossV2CrossHeight,
			},
			Ticks: 210,
		},
	}
}

func (v *V2) setBarbells() {
	v.Barbells = []*BarbellInfo{
		{
			Obj: &object.Object{
				Origin: &geometry.Point{X: v.StartPoint.X, Y: v.StartPoint.Y},
				Width:  BossV2BarbellWidth,
				Height: BossV2BarbellHeight,
			},
			StartTicks: 150,
			Ticks:      150,
		},
		{
			Obj: &object.Object{
				Origin: &geometry.Point{X: v.StartPoint.X - v.Width, Y: v.StartPoint.Y},
				Width:  BossV2BarbellWidth,
				Height: BossV2BarbellHeight,
			},
			StartTicks: 200,
			Ticks:      200,
		},
	}
}

func (v *V2) Tick(playerPos *geometry.Point) {
	v.Ticks++
	v.Health--
	if v.Health == 0 {
		v.Dead = true
		return
	}

	for _, cross := range v.Crosses {
		if cross.Tick == 0 {
			cross.Tick = cross.Ticks + cross.Ticks/10
			cross.Direction = &geometry.Vector{X: (playerPos.X - cross.Obj.Origin.X) / float64(cross.Ticks), Y: (playerPos.Y - cross.Obj.Origin.Y) / float64(cross.Ticks)}
		} else {
			cross.Tick--
			cross.Obj.Move(cross.Direction)
		}
	}

	for _, barbell := range v.Barbells {
		if barbell.StartTicks > 0 {
			barbell.StartTicks--
			continue
		}

		xc := barbell.Obj.Origin.X + barbell.Obj.Width/2

		switch barbell.State {
		case 0:
			if playerPos.X <= xc && xc <= playerPos.X+player.Width && barbell.Obj.Origin.Y < playerPos.Y {
				barbell.State = 1
			} else {
				target := playerPos.Add(&geometry.Vector{X: player.Width/2 - BossV2BarbellWidth/2, Y: -520})
				barbell.Obj.Move(&geometry.Vector{X: (target.X - barbell.Obj.Origin.X) / float64(barbell.Ticks), Y: (target.Y - barbell.Obj.Origin.Y) / float64(barbell.Ticks)})
			}
		case 1:
			barbell.State = 2
			target := (playerPos.Y + 320) - barbell.Obj.Origin.Y
			if target < 0 {
				target = 0
			}
			barbell.FallSpeed = target / 180
			barbell.FallTicks = 180
		case 2:
			barbell.FallTicks--
			if barbell.FallTicks == 0 {
				barbell.State = 0
			} else {
				barbell.Obj.Move(&geometry.Vector{Y: barbell.FallSpeed})
			}
		default:
		}
	}
}

func (v *V2) Reset() {
	v.Ticks = 0
	v.Health = v.StartHealth
	v.MoveTo(v.StartPoint)
	v.setCrosses()
	v.setBarbells()
}

func (v *V2) CreateBullets(playerPos *geometry.Point) []*damage.Bullet {
	if v.Ticks%3 != 0 {
		return nil
	}

	var bullets []*damage.Bullet

	for _, c := range []float64{30, 55, 60, 80, 100, 110} {
		bullets = append(bullets, damage.NewBullet(
			v.Origin.Add(&geometry.Vector{X: BossV1Width / 2, Y: BossV1Height / 2}),
			v.bulletImage,
			6,
			&geometry.Vector{X: (playerPos.X - v.Origin.X) / c, Y: (playerPos.Y - v.Origin.Y) / c},
		))
	}

	return bullets
}

func (v *V2) GetNextMove() (float64, float64) {
	switch v.State {
	case 0:
		if v.Origin.X > v.StartPoint.X-v.Width {
			return v.Origin.X - v.Speed, v.Origin.Y
		} else {
			v.State = 1
			return v.Origin.X, v.Origin.Y
		}
	case 1:
		if v.Origin.Y < v.StartPoint.Y+v.Height {
			return v.Origin.X, v.Origin.Y + v.Speed
		} else {
			v.State = 2
			return v.Origin.X, v.Origin.Y
		}
	case 2:
		if v.Origin.X < v.StartPoint.X {
			return v.Origin.X + v.Speed, v.Origin.Y
		} else {
			v.State = 3
			return v.Origin.X, v.Origin.Y
		}
	case 3:
		if v.Origin.Y > v.StartPoint.Y {
			return v.Origin.X, v.Origin.Y - v.Speed
		} else {
			v.State = 0
			return v.Origin.X, v.Origin.Y
		}
	default:
	}

	return v.Origin.X, v.Origin.Y
}
