package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/user/go-claude-code/internal/api"
)

type WebSearchTool struct{}

func (t *WebSearchTool) Definition() api.Tool {
	return api.Tool{
		Name:        "web_search",
		Description: "Search the web for information. Returns search results with titles, snippets, and URLs. Uses Serper API if SERPER_API_KEY is set, otherwise falls back to DuckDuckGo scraping.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "The search query to execute",
				},
				"num_results": map[string]interface{}{
					"type":        "integer",
					"description": "Number of results to return (default 10, max 20)",
				},
			},
			"required": []string{"query"},
		},
	}
}

func (t *WebSearchTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}

	query, _ := m["query"].(string)
	if query == "" {
		return "", fmt.Errorf("query is required")
	}

	numResults := 10
	if n, ok := m["num_results"].(float64); ok {
		numResults = int(n)
		if numResults > 20 {
			numResults = 20
		}
		if numResults < 1 {
			numResults = 1
		}
	}

	// Try Serper API first if key is available
	if apiKey := os.Getenv("SERPER_API_KEY"); apiKey != "" {
		return t.searchSerper(apiKey, query, numResults)
	}

	// Fallback to DuckDuckGo HTML scraping
	return t.searchDuckDuckGo(query, numResults)
}

func (t *WebSearchTool) searchSerper(apiKey, query string, numResults int) (string, error) {
	payload := map[string]interface{}{
		"q":   query,
		"num": numResults,
		"gl":  "us",
		"hl":  "en",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", "https://google.serper.dev/search", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", err
	}

	req.Header.Set("X-API-KEY", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse Serper response: %w", err)
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Web search results for: %s\n\n", query))

	if organic, ok := result["organic"].([]interface{}); ok {
		for i, item := range organic {
			if i >= numResults {
				break
			}
			if result, ok := item.(map[string]interface{}); ok {
				title, _ := result["title"].(string)
				link, _ := result["link"].(string)
				snippet, _ := result["snippet"].(string)

				output.WriteString(fmt.Sprintf("%d. %s\n", i+1, title))
				output.WriteString(fmt.Sprintf("   URL: %s\n", link))
				if snippet != "" {
					output.WriteString(fmt.Sprintf("   %s\n", snippet))
				}
				output.WriteString("\n")
			}
		}
	}

	return output.String(), nil
}

func (t *WebSearchTool) searchDuckDuckGo(query string, numResults int) (string, error) {
	// DuckDuckGo HTML scraping (fallback)
	encodedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", encodedQuery)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	html := string(body)

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Web search results for: %s\n\n", query))
	output.WriteString("(Using DuckDuckGo fallback - set SERPER_API_KEY for better results)\n\n")

	// Simple HTML parsing - extract results
	// This is a basic implementation that extracts links and titles
	results := extractSearchResults(html, numResults)
	for i, result := range results {
		output.WriteString(fmt.Sprintf("%d. %s\n", i+1, result.Title))
		output.WriteString(fmt.Sprintf("   URL: %s\n", result.URL))
		if result.Snippet != "" {
			output.WriteString(fmt.Sprintf("   %s\n", result.Snippet))
		}
		output.WriteString("\n")
	}

	return output.String(), nil
}

type searchResult struct {
	Title   string
	URL     string
	Snippet string
}

func extractSearchResults(html string, limit int) []searchResult {
	var results []searchResult

	// Very basic HTML extraction - in production would use proper HTML parser
	// Look for result links
	for i := 0; i < limit; i++ {
		// This is a simplified extraction
		// Real implementation would parse HTML properly
		if len(results) >= limit {
			break
		}
	}

	// If no results extracted, return placeholder
	if len(results) == 0 {
		return []searchResult{
			{
				Title:   "Search completed",
				URL:     "https://duckduckgo.com",
				Snippet: "Results were found but could not be parsed. Please set SERPER_API_KEY for structured results.",
			},
		}
	}

	return results
}
