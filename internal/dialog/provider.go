package dialog

import "fmt"

type Provider interface {
	Get(id string) (Dialog, error)
}

type StandardProvider struct {
}

func (sp *StandardProvider) Get(id string) (Dialog, error) {
	switch id {
	case "slonser-web-task":
		return NewDummy("Hello, I'm SLONIK! pröööh об этом", "wddsd"), nil
	default:
		return nil, fmt.Errorf("unknown dialog id: %s", id)
	}
}

type ClientProvider struct {
}

func (cp *ClientProvider) Get(_ string) (Dialog, error) {
	return NewClientDialog(), nil
}
