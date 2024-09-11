package anthropic

import (
	"context"
	"errors"
	"fmt"
)

type Service struct {
	client *Client
}

func NewService(apiKey string) *Service {
	return &Service{
		client: NewClient(apiKey),
	}
}

func (s *Service) ProcessNaturalLanguage(ctx context.Context, input string) (string, error) {
	response, err := s.client.CompletePromptJson(ctx, input, "claude-3-5-sonnet-20240620")
	if err != nil {
		return "", fmt.Errorf("failed to process natural language: %w", err)
	}

	if response.Error != "" {
		return "", errors.New(response.Error)
	}

	if response.Cron == "" {
		return "", errors.New("generated cron_internal expression is empty")
	}

	return response.Cron, nil
}
