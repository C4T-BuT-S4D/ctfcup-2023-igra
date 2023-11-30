package sprites

import (
	"fmt"
	"sync"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/resources"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	_ "image/png"
)

type Manager struct {
	cache map[string]*ebiten.Image
	m     sync.Mutex
}

func (m *Manager) getSprite(path string) (*ebiten.Image, error) {
	m.m.Lock()
	defer m.m.Unlock()

	if sprite, ok := m.cache[path]; ok {
		return sprite, nil
	}

	eimg, _, err := ebitenutil.NewImageFromFileSystem(resources.EmbeddedFS, path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sprite ('%v'): %w", path, err)
	}

	m.cache[path] = eimg
	return eimg, nil
}

func (m *Manager) GetSprite(spriteType Type) (*ebiten.Image, error) {
	return m.getSprite(fmt.Sprintf("sprites/%s.png", spriteType))
}

func (m *Manager) GetAnimationSprite(spriteType Type, animation string) (*ebiten.Image, error) {
	return m.getSprite(fmt.Sprintf("sprites/%s_%s.png", spriteType, animation))
}

func NewManager() *Manager {
	return &Manager{
		cache: make(map[string]*ebiten.Image),
	}
}
