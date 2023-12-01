package fonts

import (
	"fmt"
	"github.com/c4t-but-s4d/ctfcup-2023-igra/internal/resources"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"sync"
)

type Manager struct {
	cache map[Type]font.Face
	m     sync.Mutex
}

func NewManager() *Manager {
	return &Manager{
		cache: make(map[Type]font.Face),
	}
}

func (m *Manager) Get(t Type) font.Face {
	m.m.Lock()
	defer m.m.Unlock()

	if face, ok := m.cache[t]; ok {
		return face
	}

	f, err := resources.EmbeddedFS.ReadFile(fmt.Sprintf("fonts/%s", t))
	if err != nil {
		panic(err)
	}

	ff, err := opentype.Parse(f)
	if err != nil {
		panic(err)
	}

	face, err := opentype.NewFace(ff, &opentype.FaceOptions{
		Size:    72,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(err)
	}

	m.cache[t] = face
	return face
}
