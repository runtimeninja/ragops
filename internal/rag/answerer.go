package rag

import (
	"context"
	"strings"
)

type Answerer interface {
	Answer(ctx context.Context, model, question string, sources []Source) (string, error)
}

type OpenAIAnswerer struct {
	client *OpenAI
}

func NewOpenAIAnswerer(client *OpenAI) *OpenAIAnswerer {
	return &OpenAIAnswerer{client: client}
}

type responsesReq struct {
	Model        string `json:"model"`
	Instructions string `json:"instructions,omitempty"`
	Input        []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"input"`
}

type responsesResp struct {
	Output []struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"output"`
}

func (a *OpenAIAnswerer) Answer(ctx context.Context, model, question string, sources []Source) (string, error) {
	var ctxb strings.Builder
	for i, s := range sources {
		ctxb.WriteString("\n[")
		ctxb.WriteString(intToStr(i + 1))
		ctxb.WriteString("] doc=")
		ctxb.WriteString(s.DocumentID)
		ctxb.WriteString(" chunk=")
		ctxb.WriteString(s.ChunkID)
		ctxb.WriteString("\n")
		ctxb.WriteString(s.Content)
		ctxb.WriteString("\n")
	}

	req := responsesReq{
		Model:        model,
		Instructions: "Answer using ONLY the provided context. If not found, say you don't know. Add citations like [1], [2].",
		Input: []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{
			{Role: "user", Content: "Question:\n" + question + "\n\nContext:\n" + ctxb.String()},
		},
	}

	var resp responsesResp
	err := a.client.doJSON(ctx, "POST", "https://api.openai.com/v1/responses", req, &resp)
	if err != nil {
		return "", err
	}

	var out strings.Builder
	for _, o := range resp.Output {
		for _, c := range o.Content {
			if c.Type == "output_text" && c.Text != "" {
				out.WriteString(c.Text)
			}
		}
	}
	return strings.TrimSpace(out.String()), nil
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	var s [20]byte
	i := len(s)
	for n > 0 {
		i--
		s[i] = byte('0' + (n % 10))
		n /= 10
	}
	return string(s[i:])
}
