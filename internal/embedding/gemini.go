package embedding

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"
)

type GeminiClient struct {
	client *genai.Client
}

func NewGeminiClient(ctx context.Context) (*GeminiClient, error) {

	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &GeminiClient{
		client: client,
	}, nil
}

func (g *GeminiClient) Embed(ctx context.Context, text string) ([]float32, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	contents := []*genai.Content{
		genai.NewContentFromText(text, genai.RoleUser),
	}

	result, err := g.client.Models.EmbedContent(
		ctx,
		"gemini-embedding-2",
		contents,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("generate embedding: %w", err)
	}

	if result == nil || len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return result.Embeddings[0].Values, nil
}
