package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type OpenAI struct {
	APIKey string
	HTTP   *http.Client
}

func NewOpenAI(apiKey string) *OpenAI {
	return &OpenAI{
		APIKey: apiKey,
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *OpenAI) doJSON(ctx context.Context, method, url string, reqBody any, out any) error {
	if c.APIKey == "" {
		return errors.New("OPENAI_API_KEY missing")
	}

	var buf bytes.Buffer
	if reqBody != nil {
		if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
			return err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("openai http %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
