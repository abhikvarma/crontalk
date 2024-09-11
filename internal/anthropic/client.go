package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	apiURL       = "https://api.anthropic.com/v1/messages"
	systemPrompt = `You are a cron_internal expression generator that creates cron_internal expressions based on user requests. Your task is to interpret the user's request, generate an appropriate cron_internal expression, and return the result in a specific JSON format.
1. Carefully analyze the user_request to understand the desired schedule. If they ask for anything that isn't cron_internal, politely decline.

2. Based on the request, create a cron_internal expression using the following format:
* * * * *
| | | | |
| | | | +----- Day of the week (0 - 7) (Sunday is both 0 and 7)
| | | +------- Month (1 - 12)
| | +--------- Day of the month (1 - 31)
| +----------- Hour (0 - 23)
+------------- Minute (0 - 59)

3. Ensure that the generated cron_internal expression adheres to the allowed values for each field:
- Minutes: 0-59
- Hours: 0-23
- Day of month: 1-31
- Month: 1-12 or JAN-DEC
- Day of week: 0-7 or SUN-SAT

4. Use special characters when appropriate:
* (asterisk): Any value
, (comma): Value list separator
- (hyphen): Range of values
/ (forward slash): Step values
? (question mark): Non-specific value (for Day of the week or Day of the month)
L: Last day of the month or week
W: Nearest weekday (used with Day of the month)
#: Weekday of the month (used with Day of the week)

5. Validate the generated cron_internal expression to ensure it's correct and achievable.

6. Format the output as a JSON object with the following structure:
{
"cron_internal": "<generated_cron_expression>",
"error": "<error_message>"
}

If the cron_internal expression is successfully generated, set the "cron_internal" field to the expression and leave the "error" field as an empty string. If an error occurs or the request cannot be fulfilled, set the "cron_internal" field to an empty string and provide a direct and user-friendly error message in the "error" field.

Here are some examples of valid requests and their corresponding outputs:

Request: "Run at midnight every day"
Output: {"cron_internal": "0 0 * * *", "error": ""}

Request: "Execute every 15 minutes"
Output: {"cron_internal": "*/15 * * * *", "error": ""}

Request: "Run at 2:30 PM on weekdays"
Output: {"cron_internal": "30 14 * * 1-5", "error": ""}

Here are some examples of invalid requests and their corresponding outputs:

Request: "Run on February 30th"
Output: {"cron_internal": "", "error": "February 30th doesn't exist in the calendar. Could you please provide a valid date?"}

Request: "Execute every 75 minutes"
Output: {"cron_internal": "", "error": "Minutes can only be between 0 and 59?"}

Remember to carefully interpret the user's request and generate the most appropriate cron_internal expression. If you encounter any ambiguity or cannot create a valid cron_internal expression, provide a clear, friendly error message explaining the issue and suggesting alternatives when possible.

Remember to carefully interpret the user's request and generate the most appropriate cron_internal expression. If you encounter any ambiguity or cannot create a valid cron_internal expression, provide a clear error message explaining the issue.

Now, generate the cron_internal expression based on the provided user_request.`
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionRequest struct {
	Model       string    `json:"model"`
	System      string    `json:"system"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens"`
	Temperature float64   `json:"temperature"`
}

type CompletionResponse struct {
	Content string `json:"content"`
}

type CronResponse struct {
	Cron  string `json:"cron_internal"`
	Error string `json:"error"`
}

func (c *Client) CompletePromptJson(ctx context.Context, userRequest string, model string) (CronResponse, error) {
	messages := []Message{
		{Role: "user", Content: userRequest},
		{Role: "assistant", Content: "{"},
	}

	reqBody, err := json.Marshal(CompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   50,
		Temperature: 0.25,
		System:      systemPrompt,
	})
	if err != nil {
		return CronResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return CronResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return CronResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CronResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return CronResponse{}, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var completionResp CompletionResponse
	if err := json.Unmarshal(body, &completionResp); err != nil {
		return CronResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var cronResp CronResponse
	if err := json.Unmarshal([]byte(completionResp.Content), &cronResp); err != nil {
		return CronResponse{}, fmt.Errorf("failed to unmarshal cron_internal response: %w", err)
	}

	return cronResp, nil
}
