package llmc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type Request struct {
	Model        string         `json:"model"`
	SystemPrompt string         `json:"system"`
	UserPrompt   string         `json:"prompt"`
	Options      map[string]any `json:"options"`
	Stream       bool           `json:"stream"`
}

type Response struct {
	Response string
}

type Client struct {
	host string
}

func NewClient(host string) *Client {
	return &Client{
		host: host,
	}
}

func (c *Client) MakeRequest(ctx context.Context, req *Request, logger *logrus.Entry) (*Response, error) {
	logger.Infof("chose host %q", c.host)

	req.Stream = false
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/generate", c.host),
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer func() {
		if err := httpResp.Body.Close(); err != nil {
			logrus.Errorf("error closing response body: %v", err)
		}
	}()

	var respBody Response
	if err := json.NewDecoder(httpResp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("decoding response body: %w", err)
	}
	return &respBody, nil
}
