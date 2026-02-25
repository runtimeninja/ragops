package rag

import (
	"context"

	"github.com/pgvector/pgvector-go"
)

type Embedder interface {
	Embed(ctx context.Context, text string) (pgvector.Vector, int, error)
}

type OpenAIEmbedder struct {
	client *OpenAI
	model  string
}

func NewOpenAIEmbedder(client *OpenAI, model string) *OpenAIEmbedder {
	return &OpenAIEmbedder{client: client, model: model}
}

type embeddingsReq struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embeddingsResp struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

func (e *OpenAIEmbedder) Embed(ctx context.Context, text string) (pgvector.Vector, int, error) {
	var resp embeddingsResp
	err := e.client.doJSON(ctx, "POST", "https://api.openai.com/v1/embeddings", embeddingsReq{
		Model: e.model,
		Input: text,
	}, &resp)
	if err != nil {
		return pgvector.Vector{}, 0, err
	}
	vec := pgvector.NewVector(resp.Data[0].Embedding)
	return vec, resp.Usage.TotalTokens, nil
}
