package dialog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

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

func (L *LLM) Feed(text string) error {
	if text == L.Flag {
		L.state.GaveItem = true
		L.state.Finished = true
		L.state.Text = "You have defeated me!"
		return nil
	}

	body, err := json.Marshal(map[string]any{
		"system": L.Prompt,
		"prompt": text,
	})
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	llmReq, err := http.NewRequest("POST", L.URL+"/api/generate", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	llmReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(llmReq)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	var respBody map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return fmt.Errorf("decoding response body: %w", err)
	}

	L.state.Text = respBody["response"]
	return nil
}

func (L *LLM) State() *State {
	return &L.state
}

func (L *LLM) SetState(s *State) {
	// No need for.
}
