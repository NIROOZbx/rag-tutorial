package groq

import (
	"context"
	"errors"
	"strings"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

const defaultModel = "openai/gpt-oss-20b"

type Client struct {
	client *openai.Client
	model  string
}

func NewClient(apiKey string) (*Client, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, errors.New("GROQ_API_KEY is required")
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://api.groq.com/openai/v1"),
	)

	return &Client{client: &client, model: defaultModel}, nil
}

func (c *Client) GenerateAnswer(ctx context.Context, prompt string) (string, error) {
	completion, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{openai.UserMessage(prompt)},
	})
	if err != nil {
		return "", err
	}
	if completion == nil || len(completion.Choices) == 0 {
		return "", errors.New("Groq returned no answer choices")
	}

	answer := strings.TrimSpace(completion.Choices[0].Message.Content)
	if answer == "" {
		return "", errors.New("Groq returned an empty answer")
	}
	return answer, nil
}
