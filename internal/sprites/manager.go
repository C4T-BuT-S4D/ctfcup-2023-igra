package sprites

import (
	"fmt"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/resources"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"sync"

	_ "image/png"
)

type Manager struct {
	cache map[string]*ebiten.Image
	m     sync.Mutex
}

func (m *Manager) GetSprite(spriteType Type) (*ebiten.Image, error) {
	m.m.Lock()
	defer m.m.Unlock()

	if sprite, ok := m.cache[string(spriteType)]; ok {
		return sprite, nil
	}

	path := fmt.Sprintf("sprites/%s.png", spriteType)

	eimg, _, err := ebitenutil.NewImageFromFileSystem(resources.EmbeddedFS, path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sprite ('%v'): %w", path, err)
	}

	m.cache[string(spriteType)] = eimg
	return eimg, nil
}

func NewManager() *Manager {
	return &Manager{
		cache: make(map[string]*ebiten.Image),
	}
}
