package dialog

import (
	"fmt"
	"os"
)

type Provider interface {
	Get(id string) (Dialog, error)
	DisplayInput() bool
}

type StandardProvider struct {
	ShowInput bool
}

func NewStandardProvider(showInput bool) *StandardProvider {
	return &StandardProvider{
		ShowInput: showInput,
	}
}

func (sp *StandardProvider) DisplayInput() bool {
	return sp.ShowInput
}

func (sp *StandardProvider) Get(id string) (Dialog, error) {
	switch id {
	case "test-npc":
		return NewDummy("Hello, I'm a test NPC!\n 2 + 2 = ?", "4"), nil
	case "slonser-task":
		return NewDummy("Hello, I'm SLONIK and I'm ______ об этом", "pröööh"), nil
	case "wise-tree-task":
		return NewWiseTree(), nil
	case "llm-boss-1":
		return &LLM{
			Intro:     "I'm the boss of Arbalest of Siberia. Don't try to defeat me!",
			Token:     os.Getenv("AUTH_TOKEN"),
			URL:       "http://localhost:8081",
			MaskInput: !sp.DisplayInput(),
		}, nil
	case "ceo-boss":
		return &LLM{
			Intro:     "I'm the REAL boss of Arbalest of Siberia. It's impossible to defeat me!",
			Token:     os.Getenv("AUTH_TOKEN"),
			URL:       "http://localhost:8082",
			MaskInput: !sp.DisplayInput(),
		}, nil

	case "capytoshka":
		return NewCapy(os.Getenv("CAPY_TOKEN")), nil
	default:
		return nil, fmt.Errorf("unknown dialog id: %s", id)
	}
}

type ClientProvider struct {
}

func (cp *ClientProvider) Get(_ string) (Dialog, error) {
	return NewClientDialog(), nil
}

func (cp *ClientProvider) DisplayInput() bool {
	return true
}
