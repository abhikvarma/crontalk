package anthropic

import (
	"context"
	"errors"
	"fmt"
)

type Service struct {
	client *Client
}

func NewService(apiKey, model string) *Service {
	return &Service{
		client: NewClient(apiKey, model),
	}
}

func (s *Service) ProcessCronQuestion(ctx context.Context, input string) (LlmCronResponse, error) {
	cronResp, err := s.client.CompletePromptJson(ctx, input)
	if err != nil {
		return LlmCronResponse{}, fmt.Errorf("failed to process cron question: %w", err)
	}

	if cronResp.Cron == "" && cronResp.Error == "" {
		return LlmCronResponse{}, errors.New("generated cron expression is empty")
	}

	return cronResp, nil
}
