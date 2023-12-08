package dialog

import "fmt"

type Provider interface {
	Get(id string) (Dialog, error)
}

type StandardProvider struct {
}

func (sp *StandardProvider) Get(id string) (Dialog, error) {
	switch id {
	case "test-npc":
		return NewDummy("Hello, I'm a test NPC!\n 2 + 2 = ?", "4"), nil
	case "wise-tree-task":
		return NewWiseTree(), nil
	default:
		return nil, fmt.Errorf("unknown dialog id: %s", id)
	}
}

type ClientProvider struct {
}

func (cp *ClientProvider) Get(_ string) (Dialog, error) {
	return NewClientDialog(), nil
}
