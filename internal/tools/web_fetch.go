package tools

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/user/go-claude-code/internal/api"
	"github.com/yuin/goldmark"
)

type WebFetchTool struct{}

func (t *WebFetchTool) Definition() api.Tool {
	return api.Tool{
		Name:        "web_fetch",
		Description: "Fetch and extract content from a URL. Returns the extracted text content, removing HTML tags, scripts, and styles. Useful for reading web pages, documentation, or articles.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "The URL to fetch",
				},
				"max_length": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum characters to return (default 10000, max 50000)",
				},
			},
			"required": []string{"url"},
		},
	}
}

func (t *WebFetchTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}

	urlStr, _ := m["url"].(string)
	if urlStr == "" {
		return "", fmt.Errorf("url is required")
	}

	// Validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", fmt.Errorf("only http and https URLs are supported")
	}

	maxLength := 10000
	if ml, ok := m["max_length"].(float64); ok {
		maxLength = int(ml)
		if maxLength > 50000 {
			maxLength = 50000
		}
	}

	// Fetch the URL
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Read body
	body, err := io.ReadAll(io.LimitReader(resp.Body, int64(maxLength*2)))
	if err != nil {
		return "", err
	}

	contentType := resp.Header.Get("Content-Type")
	content := string(body)

	// Extract text based on content type
	var textContent string
	if strings.Contains(contentType, "text/html") || strings.Contains(content, "<html") {
		textContent = extractTextFromHTML(content)
	} else {
		textContent = content
	}

	// Truncate if needed
	if len(textContent) > maxLength {
		textContent = textContent[:maxLength] + "\n\n[Content truncated...]"
	}

	var output strings.Builder
	output.WriteString(fmt.Sprintf("Content from: %s\n", urlStr))
	output.WriteString(fmt.Sprintf("Title: %s\n\n", extractTitle(content)))
	output.WriteString(textContent)

	return output.String(), nil
}

func extractTextFromHTML(html string) string {
	// Remove script and style tags
	html = removeTag(html, "script")
	html = removeTag(html, "style")
	html = removeTag(html, "nav")
	html = removeTag(html, "header")
	html = removeTag(html, "footer")
	html = removeTag(html, "aside")

	// Convert to markdown then plain text
	var buf strings.Builder
	if err := goldmark.Convert([]byte(html), &buf); err == nil {
		return cleanWhitespace(buf.String())
	}

	// Fallback: simple tag stripping
	return simpleStripTags(html)
}

func removeTag(html, tag string) string {
	// Simple regex-like removal
	startTag := "<" + tag
	endTag := "</" + tag + ">"

	for {
		startIdx := strings.Index(strings.ToLower(html), startTag)
		if startIdx == -1 {
			break
		}

		// Find end of opening tag
		endOpenIdx := strings.Index(html[startIdx:], ">")
		if endOpenIdx == -1 {
			break
		}
		endOpenIdx += startIdx + 1

		// Find closing tag
		endCloseIdx := strings.Index(strings.ToLower(html[endOpenIdx:]), endTag)
		if endCloseIdx == -1 {
			// Self-closing or no closing tag - just remove the opening tag
			html = html[:startIdx] + html[endOpenIdx:]
			continue
		}
		endCloseIdx += endOpenIdx + len(endTag)

		html = html[:startIdx] + " " + html[endCloseIdx:]
	}

	return html
}

func simpleStripTags(html string) string {
	var result strings.Builder
	inTag := false

	for _, r := range html {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
			result.WriteByte(' ')
		default:
			if !inTag {
				result.WriteRune(r)
			}
		}
	}

	return cleanWhitespace(result.String())
}

func extractTitle(html string) string {
	lowerHTML := strings.ToLower(html)
	startIdx := strings.Index(lowerHTML, "<title>")
	if startIdx == -1 {
		return "No title found"
	}
	startIdx += 7

	endIdx := strings.Index(lowerHTML[startIdx:], "</title>")
	if endIdx == -1 {
		return "No title found"
	}

	return strings.TrimSpace(html[startIdx : startIdx+endIdx])
}

func cleanWhitespace(s string) string {
	// Collapse multiple whitespace
	var result strings.Builder
	lastWasSpace := true

	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if !lastWasSpace {
				result.WriteByte(' ')
				lastWasSpace = true
			}
		} else {
			result.WriteRune(r)
			lastWasSpace = false
		}
	}

	return strings.TrimSpace(result.String())
}
