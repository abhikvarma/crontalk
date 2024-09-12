package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const (
	apiURL       = "https://api.anthropic.com/v1/messages"
	systemPrompt = `You are a cron expression generator that creates cron expressions based on user requests. Your task is to interpret the user's request, generate an appropriate cron expression, and return the result in a specific JSON format.
1. Carefully analyze the user_request to understand the desired schedule. If they ask for anything that isn't cron, politely decline.

2. Based on the request, create a cron expression using the following format:
* * * * *
| | | | |
| | | | +----- Day of the week (0 - 7) (Sunday is both 0 and 7)
| | | +------- Month (1 - 12)
| | +--------- Day of the month (1 - 31)
| +----------- Hour (0 - 23)
+------------- Minute (0 - 59)

3. Ensure that the generated cron expression adheres to the allowed values for each field:
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

5. Validate the generated cron expression to ensure it's correct and achievable.

6. Format the output as a JSON object with the following structure:
{
"cron": "<generated_cron_expression>",
"error": "<error_message>"
}

If the cron expression is successfully generated, set the "cron" field to the expression and leave the "error" field as an empty string. 
If an error occurs or the request cannot be fulfilled, set the "cron" field to an empty string and provide an educational error message in the "error" field (max 20 words).
The error message should contain which field is wrong and why. Suggest potential alternatives when possible.

Here are some examples of valid requests and their corresponding outputs:

Request: "Run at midnight every day"
Output: {"cron": "0 0 * * *", "error": ""}

Request: "Execute every 15 minutes"
Output: {"cron": "*/15 * * * *", "error": ""}

Request: "Run at 2:30 PM on weekdays"
Output: {"cron": "30 14 * * 1-5", "error": ""}

Here are some examples of invalid requests and their corresponding outputs:

Request: "Run on February 30th"
Output: {"cron": "", "error": "February 30th doesn't exist in the calendar. Try using another date"}

Request: "Execute every 75 minutes"
Output: {"cron": "", "error": "Minutes can only be between 0 and 59"}

Remember to carefully interpret the user's request and generate the most appropriate cron expression. If you encounter any ambiguity or cannot create a valid cron expression, provide a clear, friendly error message explaining the issue and suggesting alternatives when possible.

Remember to carefully interpret the user's request and generate the most appropriate cron expression. If you encounter any ambiguity or cannot create a valid cron expression, provide a clear error message explaining the issue.

Now, generate the cron expression based on the provided user_request.`
)

type Client struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey:     apiKey,
		model:      model,
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
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

type LlmCronResponse struct {
	Cron  string `json:"cron"`
	Error string `json:"error"`
}

func (c *Client) CompletePromptJson(ctx context.Context, userRequest string) (LlmCronResponse, error) {
	messages := []Message{
		{Role: "user", Content: userRequest},
		{Role: "assistant", Content: "{"},
	}

	reqBody, err := json.Marshal(CompletionRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   300,
		Temperature: 0.25,
		System:      systemPrompt,
	})
	if err != nil {
		return LlmCronResponse{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return LlmCronResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return LlmCronResponse{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return LlmCronResponse{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return LlmCronResponse{}, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var completionResp CompletionResponse
	if err := json.Unmarshal(body, &completionResp); err != nil {
		return LlmCronResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(completionResp.Content) == 0 {
		return LlmCronResponse{}, fmt.Errorf("empty response content")
	}

	log.Printf("Received response %v", completionResp)
	jsonStr := completionResp.Content[0].Text
	if !strings.HasPrefix(strings.TrimSpace(jsonStr), "{") {
		jsonStr = "{" + jsonStr
	}

	var cronResp LlmCronResponse
	if err := json.Unmarshal([]byte(jsonStr), &cronResp); err != nil {
		return LlmCronResponse{}, fmt.Errorf("failed to unmarshal cron response: %w", err)
	}

	return cronResp, nil
}
