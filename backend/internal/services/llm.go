package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// LLMService provides an interface for generating interview instructions from rubrics
type LLMService interface {
	GenerateInterviewInstructions(ctx context.Context, rubricTitle string, rubricRawText string) (string, error)
}

// OpenAIService implements LLMService using OpenAI API (or Anthropic as fallback)
type OpenAIService struct {
	apiKey  string
	baseURL string
	client  *http.Client
	useAnthropic bool
}

// NewOpenAIService creates a new OpenAI service instance
func NewOpenAIService() *OpenAIService {
	openAIKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	
	var apiKey string
	var useAnthropic bool
	
	if openAIKey != "" {
		apiKey = openAIKey
		useAnthropic = false
	} else if anthropicKey != "" {
		apiKey = anthropicKey
		useAnthropic = true
	}

	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &OpenAIService{
		apiKey:  apiKey,
		baseURL: baseURL,
		useAnthropic: useAnthropic,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// GenerateInterviewInstructions generates interview instructions from a rubric using the LLM
func (s *OpenAIService) GenerateInterviewInstructions(ctx context.Context, rubricTitle string, rubricRawText string) (string, error) {
	if s.apiKey == "" {
		return "", fmt.Errorf("LLM API key not configured (set OPENAI_API_KEY or ANTHROPIC_API_KEY)")
	}

	// Build the prompt for the LLM
	prompt := fmt.Sprintf(`You are an expert educational assessment designer. Based on the following rubric, generate comprehensive interview instructions that will guide an AI interviewer to conduct an effective assessment interview.

Rubric Title: %s

Rubric Content:
%s

Please generate detailed interview instructions that:
1. Explain the purpose and goals of the interview
2. Provide guidance on how to assess the student's understanding
3. Include instructions for probing deeper when needed
4. Guide the interviewer on how to evaluate responses against the rubric criteria
5. Suggest appropriate follow-up questions based on student responses

Return only the instructions text, without any additional formatting or explanations.`, rubricTitle, rubricRawText)

	if s.useAnthropic {
		return s.callAnthropicAPI(ctx, prompt)
	}

	return s.callOpenAIAPI(ctx, prompt)
}

// callOpenAIAPI makes a request to OpenAI's API
func (s *OpenAIService) callOpenAIAPI(ctx context.Context, prompt string) (string, error) {
	url := fmt.Sprintf("%s/chat/completions", s.baseURL)

	requestBody := map[string]interface{}{
		"model": "gpt-4o-mini", // Can be made configurable via env var
		"messages": []map[string]string{
			{
				"role":    "system",
				"content": "You are an expert educational assessment designer. Generate clear, detailed interview instructions based on rubrics.",
			},
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
		"max_tokens":  2000,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.apiKey))

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

// callAnthropicAPI makes a request to Anthropic's Claude API
func (s *OpenAIService) callAnthropicAPI(ctx context.Context, prompt string) (string, error) {
	url := "https://api.anthropic.com/v1/messages"

	requestBody := map[string]interface{}{
		"model": "claude-3-5-sonnet-20241022", // Can be made configurable via env var
		"max_tokens": 2000,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Content) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return response.Content[0].Text, nil
}
