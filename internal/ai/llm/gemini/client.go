package gemini

import (
	"context"

	"google.golang.org/genai"
)

type GeminiFlash struct {
	client *genai.Client
	cfg    *Config
}

func NewGeminiFlashClient(ctx context.Context, cfg *Config) (*GeminiFlash, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.APIKey,
		Backend: genai.BackendGeminiAPI})

	if err != nil {
		return nil, err
	}

	return &GeminiFlash{
		client: client,
		cfg:    cfg,
	}, nil
}

func (g *GeminiFlash) TestModel(ctx context.Context) (string, error) {
	result, err := g.client.Models.GenerateContent(ctx, g.cfg.Model, genai.Text("What is the capital of France?"), &genai.GenerateContentConfig{
		Temperature: g.cfg.Temperature,
	})

	if err != nil {
		return "", err
	}

	return result.Text(), nil
}
