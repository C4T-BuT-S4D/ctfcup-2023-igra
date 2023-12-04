package dialog

import "fmt"

type Provider interface {
	Get(id string) (Dialog, error)
}

type StandardProvider struct {
}

func (sp *StandardProvider) Get(id string) (Dialog, error) {
	switch id {
	case "slonser-task":
		return NewDummy("Hello, I'm SLONIK! pröööh об этом", "pröööh"), nil
	case "wise-tree-task":
		return NewDummy("I'm wise tree", "slavsarethebest"), nil
	default:
		return nil, fmt.Errorf("unknown dialog id: %s", id)
	}
}

type ClientProvider struct {
}

func (cp *ClientProvider) Get(_ string) (Dialog, error) {
	return NewClientDialog(), nil
}
