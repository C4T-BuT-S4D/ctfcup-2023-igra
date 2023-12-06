package llmc

import (
	"context"

	"github.com/samber/lo"

	"github.com/c4t-but-s4d/ctfcup-2023-igra/pkg/sync"

	"go.uber.org/atomic"
)

type Manager struct {
	hosts []string
	locks []*sync.CtxMutex

	counter *atomic.Int64
}

func NewManager(hosts []string) *Manager {
	return &Manager{
		hosts: hosts,
		locks: lo.Map(
			hosts,
			func(_ string, _ int) *sync.CtxMutex {
				return sync.NewCtxMutex()
			},
		),
		counter: atomic.NewInt64(0),
	}
}

func (m *Manager) Acquire(ctx context.Context) (client *Client, release func()) {
	hostIndex := m.counter.Inc() % int64(len(m.hosts))
	if !m.locks[hostIndex].Lock(ctx) {
		return nil, func() {}
	}
	return NewClient(m.hosts[hostIndex]), func() {
		m.locks[hostIndex].Unlock()
	}
}
