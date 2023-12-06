package dialog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const timeout time.Duration = 30 * time.Second

type LLM struct {
	Prompt string
	Flag   string
	Intro  string
	URL    string
	state  State
}

func (L *LLM) Greeting() {
	L.state.Text = L.Intro
}

func (L *LLM) Feed(text string) {
	if text == L.Flag {
		L.state.GaveItem = true
		L.state.Finished = true
		L.state.Text += "\nYou have defeated me!"
		return
	}

	body, err := json.Marshal(map[string]any{
		"system": L.Prompt,
		"prompt": text,
	})
	if err != nil {
		L.state.Text += fmt.Sprintf("\nError: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	llmReq, err := http.NewRequestWithContext(ctx, "POST", L.URL+"/api/generate", bytes.NewBuffer(body))
	if err != nil {
		L.state.Text += fmt.Sprintf("\nError: %v", err)
		return
	}
	llmReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(llmReq)
	if err != nil {
		L.state.Text += fmt.Sprintf("\nError: %v", err)
		return
	}
	defer resp.Body.Close()

	var respBody map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		L.state.Text += fmt.Sprintf("\nError: %v", err)
		return
	}

	L.state.Text = respBody["response"]
	return
}

func (L *LLM) State() *State {
	return &L.state
}

func (L *LLM) SetState(s *State) {
	// No need for.
}
