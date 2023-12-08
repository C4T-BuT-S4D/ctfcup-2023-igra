package dialog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const timeout time.Duration = 30 * time.Second

type LLM struct {
	Intro     string
	URL       string
	Token     string
	MaskInput bool
	state     State
}

func (l *LLM) Greeting() {
	l.state.Text = l.Intro
}

func (l *LLM) callProxy(url string, body []byte) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	llmReq, err := http.NewRequestWithContext(ctx, "POST", l.URL+url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	llmReq.Header.Set("Content-Type", "application/json")
	llmReq.Header.Set("X-Team", l.Token)

	resp, err := http.DefaultClient.Do(llmReq)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	return resp, nil
}

func (l *LLM) checkIsFlag(text string) bool {
	body, err := json.Marshal(map[string]string{
		"password": text,
	})
	if err != nil {
		l.state.Text += fmt.Sprintf("Error: %v\n", err)
		return false
	}

	resp, err := l.callProxy("/api/check_password", body)
	if err != nil {
		l.state.Text += fmt.Sprintf("Error: %v\n", err)
		return false
	}

	defer func() {
		_ = resp.Body.Close()
	}()
	return resp.StatusCode == http.StatusOK
}

func (l *LLM) Feed(text string) {
	l.state.Text = fmt.Sprintf("> %s\n", text)
	if l.MaskInput {
		l.state.Text = fmt.Sprintf("> %s\n", strings.Repeat("*", len(text)))
	}

	if l.checkIsFlag(text) {
		l.state.Text += "You defeated me!!!\n"
		l.state.GaveItem = true
		l.state.Finished = true
		return
	}

	body, err := json.Marshal(map[string]string{
		"prompt": text,
	})
	if err != nil {
		l.state.Text += fmt.Sprintf("Error: %v\n", err)
		return
	}

	resp, err := l.callProxy("/api/generate", body)
	if err != nil {
		l.state.Text += fmt.Sprintf("Error: %v\n", err)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	var respBody map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		l.state.Text += fmt.Sprintf("Error: %v\n", err)
		return
	}

	l.state.Text += fmt.Sprintf("%s\n", respBody["response"])
}

func (l *LLM) State() *State {
	return &l.state
}

func (l *LLM) SetState(_ *State) {
	// No need for.
}
