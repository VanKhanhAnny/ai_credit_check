package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
	"bytes"
	"sync"
)

const (
	geminiTimeout = 600 * time.Second // Increased to 10 minutes for very slow responses
	// Free tier allows 2 requests per minute, so we wait 35 seconds between requests to be safe
	geminiRateLimitDelay = 35 * time.Second
)

var (
	lastGeminiRequest time.Time
	geminiMutex       sync.Mutex
)

// GeminiClient handles communication with the Google Gemini API
type GeminiClient struct {
	apiKey string
	model  string
}

// NewGeminiClient creates a new Gemini client
func NewGeminiClient() (*GeminiClient, error) {
	apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY"))
	if apiKey == "" {
		return nil, errors.New("GEMINI_API_KEY is not set; set it in your environment or .env")
	}

	// Default to Gemini Pro
	model := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	if model == "" {
		model = "gemini-2.5-pro" // Current model as of 2024
	}

	return &GeminiClient{
		apiKey: apiKey,
		model:  model,
	}, nil
}

// GeminiRequest represents a request to the Gemini API
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

// GeminiContent represents content in a Gemini request
type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

// GeminiPart represents a part in Gemini content
type GeminiPart struct {
	Text string `json:"text"`
}

// GeminiResponse represents a response from the Gemini API
type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

// enforceRateLimit ensures we don't exceed Gemini free tier limits
func enforceRateLimit() {
	geminiMutex.Lock()
	defer geminiMutex.Unlock()
	
	now := time.Now()
	timeSinceLastRequest := now.Sub(lastGeminiRequest)
	
	if timeSinceLastRequest < geminiRateLimitDelay {
		sleepDuration := geminiRateLimitDelay - timeSinceLastRequest
		fmt.Printf("Rate limiting: waiting %v before next Gemini request...\n", sleepDuration.Round(time.Second))
		time.Sleep(sleepDuration)
	}
	
	lastGeminiRequest = time.Now()
}

// AnalyzeDocument analyzes a document using Gemini to extract relevant information
func (c *GeminiClient) AnalyzeDocument(ctx context.Context, text string, source DocumentSource) (map[string]interface{}, error) {
	return c.analyzeDocumentWithRetry(ctx, text, source, 0)
}

// analyzeDocumentWithRetry handles the actual analysis with retry logic
func (c *GeminiClient) analyzeDocumentWithRetry(ctx context.Context, text string, source DocumentSource, retryCount int) (map[string]interface{}, error) {
	// Enforce rate limiting for free tier
	enforceRateLimit()
	
	prompt := generatePromptForSource(text, source)
	
	// Combine system instructions with the user prompt since Gemini doesn't support system role
	combinedPrompt := "You are an AI assistant that extracts structured information from documents.\n\n" + prompt
	
	req := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: combinedPrompt},
				},
				Role: "user",
			},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpCtx, cancel := context.WithTimeout(ctx, geminiTimeout)
	defer cancel()

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1/models/%s:generateContent?key=%s", c.model, c.apiKey)
	httpReq, err := http.NewRequestWithContext(httpCtx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build http request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: geminiTimeout}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		
		// Handle 429 rate limit errors with retry
		if resp.StatusCode == 429 {
			// Parse the retry delay from the response
			var errorResp struct {
				Error struct {
					Details []struct {
						RetryInfo struct {
							RetryDelay string `json:"retryDelay"`
						} `json:"retryInfo"`
					} `json:"details"`
				} `json:"error"`
			}
			
			if json.Unmarshal(respBody, &errorResp) == nil && len(errorResp.Error.Details) > 0 {
				if retryDelay := errorResp.Error.Details[0].RetryInfo.RetryDelay; retryDelay != "" {
					if duration, err := time.ParseDuration(retryDelay); err == nil {
						fmt.Printf("Rate limit hit, waiting %v before retry...\n", duration)
						time.Sleep(duration)
						// Update the last request time to account for the wait
						geminiMutex.Lock()
						lastGeminiRequest = time.Now()
						geminiMutex.Unlock()
						// Retry the request
						return c.analyzeDocumentWithRetry(ctx, text, source, retryCount+1)
					}
				}
			}
		}
		
		// Handle 503 Service Unavailable errors with exponential backoff retry
		if resp.StatusCode == 503 {
			if retryCount < 3 { // Max 3 retries for 503 errors
				fmt.Printf("Service unavailable (503), retrying (attempt %d/3)...\n", retryCount+1)
				// Wait with exponential backoff: 5s, 10s, 20s
				retryDelay := time.Duration(5*(1<<retryCount)) * time.Second
				time.Sleep(retryDelay)
				// Update the last request time to account for the wait
				geminiMutex.Lock()
				lastGeminiRequest = time.Now()
				geminiMutex.Unlock()
				// Retry the request
				return c.analyzeDocumentWithRetry(ctx, text, source, retryCount+1)
			} else {
				fmt.Printf("Service unavailable (503), max retries exceeded\n")
			}
		}
		
		return nil, fmt.Errorf("gemini http error: %s - %s", resp.Status, string(respBody))
	}

	var geminiResp GeminiResponse
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if geminiResp.Error.Message != "" {
		return nil, fmt.Errorf("gemini error: %s", geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("gemini: empty response")
	}

	// Parse the JSON response
	content := geminiResp.Candidates[0].Content.Parts[0].Text
	
	// Extract JSON from the response (it might be wrapped in markdown code blocks)
	jsonStr := extractJSONFromGemini(content)
	if jsonStr == "" {
		return nil, fmt.Errorf("could not extract JSON from response: %s", content)
	}

	// Try to unmarshal as an object first
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		// If it fails, try to unmarshal as an array and convert to object
		var arr []interface{}
		if err := json.Unmarshal([]byte(jsonStr), &arr); err != nil {
			return nil, fmt.Errorf("unmarshal response: %w", err)
		}
		// Convert array to object by using index as key
		result = make(map[string]interface{})
		for i, item := range arr {
			result[fmt.Sprintf("item_%d", i)] = item
		}
	}

	return result, nil
}

// extractJSONFromGemini extracts JSON from a string that might contain markdown
func extractJSONFromGemini(content string) string {
	// Remove markdown code blocks if present
	content = strings.TrimSpace(content)
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
	}
	
	// Find the first { or [ character
	start := strings.IndexAny(content, "{[")
	if start == -1 {
		return ""
	}
	
	// Find the matching closing character
	var end int
	var openChar, closeChar byte
	if content[start] == '{' {
		openChar, closeChar = '{', '}'
	} else {
		openChar, closeChar = '[', ']'
	}
	
	openCount := 0
	for i := start; i < len(content); i++ {
		if content[i] == openChar {
			openCount++
		} else if content[i] == closeChar {
			openCount--
			if openCount == 0 {
				end = i + 1
				break
			}
		}
	}
	
	if end == 0 {
		return ""
	}
	
	return strings.TrimSpace(content[start:end])
}
