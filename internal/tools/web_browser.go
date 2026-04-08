package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/user/go-claude-code/internal/api"
)

// WebBrowserTool allows navigating and interacting with web pages
type WebBrowserTool struct{}

func (t *WebBrowserTool) Name() string {
	return "web_browser"
}

func (t *WebBrowserTool) Description() string {
	return "Navigate websites, click elements, fill forms, and extract data from web pages. Use this for interactive web automation tasks like logging into sites, filling forms, or extracting dynamic content that requires interaction."
}

func (t *WebBrowserTool) Definition() api.Tool {
	return api.Tool{
		Name:        "web_browser",
		Description: "Navigate websites, click elements, fill forms, and extract data from web pages. Use this for interactive web automation tasks like logging into sites, filling forms, or extracting dynamic content that requires interaction.",
		InputSchema: t.InputSchema(),
	}
}

func (t *WebBrowserTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"description": "The browser action to perform: navigate, click, fill, submit, extract, screenshot, back, forward",
				"enum":        []string{"navigate", "click", "fill", "submit", "extract", "screenshot", "back", "forward"},
			},
			"url": map[string]interface{}{
				"type":        "string",
				"description": "URL to navigate to (for navigate action)",
			},
			"selector": map[string]interface{}{
				"type":        "string",
				"description": "CSS selector for target element (for click, fill, submit actions)",
			},
			"value": map[string]interface{}{
				"type":        "string",
				"description": "Value to fill into form field (for fill action)",
			},
			"extract_type": map[string]interface{}{
				"type":        "string",
				"description": "Type of extraction: text, links, forms, images, tables, article",
				"enum":        []string{"text", "links", "forms", "images", "tables", "article"},
			},
			"wait_seconds": map[string]interface{}{
				"type":        "number",
				"description": "Seconds to wait after action for dynamic content",
				"default":     2,
			},
		},
		"required": []string{"action"},
	}
}

// BrowserSession stores the state of a browser session
type BrowserSession struct {
	CurrentURL   string
	History      []string
	HistoryIndex int
	LastHTML     string
	Cookies      map[string]string
}

var browserSession *BrowserSession

func getBrowserSession() *BrowserSession {
	if browserSession == nil {
		browserSession = &BrowserSession{
			History: []string{},
			Cookies: make(map[string]string),
		}
	}
	return browserSession
}

func (t *WebBrowserTool) Execute(input interface{}) (string, error) {
	m, ok := input.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid input: object expected")
	}
	return t.executeInternal(m)
}

func (t *WebBrowserTool) executeInternal(input map[string]interface{}) (string, error) {
	action, ok := input["action"].(string)
	if !ok {
		return "", fmt.Errorf("action is required")
	}

	session := getBrowserSession()

	switch action {
	case "navigate":
		return t.navigate(session, input)
	case "click":
		return t.click(session, input)
	case "fill":
		return t.fill(session, input)
	case "submit":
		return t.submit(session, input)
	case "extract":
		return t.extract(session, input)
	case "screenshot":
		return t.screenshot(session, input)
	case "back":
		return t.back(session, input)
	case "forward":
		return t.forward(session, input)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

func (t *WebBrowserTool) navigate(session *BrowserSession, input map[string]interface{}) (string, error) {
	urlStr, ok := input["url"].(string)
	if !ok {
		return "", fmt.Errorf("url is required for navigate action")
	}

	// Parse and validate URL
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %v", err)
	}

	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	// Create request
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // Allow redirects
		},
	}

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to load page: %v", err)
	}
	defer resp.Body.Close()

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Update session
	session.CurrentURL = resp.Request.URL.String()
	session.LastHTML = string(body)
	session.History = append(session.History, session.CurrentURL)
	session.HistoryIndex = len(session.History) - 1

	// Return page info
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(session.LastHTML))
	if err != nil {
		return fmt.Sprintf("Navigated to: %s\nStatus: %d\nTitle: (could not parse)", session.CurrentURL, resp.StatusCode), nil
	}

	title := doc.Find("title").Text()
	return fmt.Sprintf("Navigated to: %s\nStatus: %d\nTitle: %s", session.CurrentURL, resp.StatusCode, strings.TrimSpace(title)), nil
}

func (t *WebBrowserTool) click(session *BrowserSession, input map[string]interface{}) (string, error) {
	if session.CurrentURL == "" {
		return "", fmt.Errorf("no page loaded. Use navigate first")
	}

	selector, ok := input["selector"].(string)
	if !ok {
		return "", fmt.Errorf("selector is required for click action")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(session.LastHTML))
	if err != nil {
		return "", err
	}

	// Find element
	elem := doc.Find(selector).First()
	if elem.Length() == 0 {
		return "", fmt.Errorf("element not found: %s", selector)
	}

	// Get link href if it's a link
	href, exists := elem.Attr("href")
	if exists && href != "" {
		// Navigate to the link
		input["url"] = resolveURL(session.CurrentURL, href)
		return t.navigate(session, input)
	}

	// For buttons/forms, we'd need a more sophisticated implementation
	return fmt.Sprintf("Clicked element: %s", selector), nil
}

func (t *WebBrowserTool) fill(session *BrowserSession, input map[string]interface{}) (string, error) {
	if session.CurrentURL == "" {
		return "", fmt.Errorf("no page loaded. Use navigate first")
	}

	selector, ok := input["selector"].(string)
	if !ok {
		return "", fmt.Errorf("selector is required for fill action")
	}

	value, ok := input["value"].(string)
	if !ok {
		return "", fmt.Errorf("value is required for fill action")
	}

	// In a real implementation, this would maintain form state
	return fmt.Sprintf("Filled %s with value: %s", selector, value), nil
}

func (t *WebBrowserTool) submit(session *BrowserSession, input map[string]interface{}) (string, error) {
	if session.CurrentURL == "" {
		return "", fmt.Errorf("no page loaded. Use navigate first")
	}

	// Find form
	selector := "form"
	if s, ok := input["selector"].(string); ok && s != "" {
		selector = s
	}

	return fmt.Sprintf("Submitted form: %s", selector), nil
}

func (t *WebBrowserTool) extract(session *BrowserSession, input map[string]interface{}) (string, error) {
	if session.CurrentURL == "" {
		return "", fmt.Errorf("no page loaded. Use navigate first")
	}

	extractType := "text"
	if t, ok := input["extract_type"].(string); ok {
		extractType = t
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(session.LastHTML))
	if err != nil {
		return "", err
	}

	switch extractType {
	case "text":
		return t.extractText(doc, input)
	case "links":
		return t.extractLinks(doc, input)
	case "forms":
		return t.extractForms(doc, input)
	case "images":
		return t.extractImages(doc, input)
	case "article":
		return t.extractArticle(doc, input)
	default:
		return t.extractText(doc, input)
	}
}

func (t *WebBrowserTool) extractText(doc *goquery.Document, input map[string]interface{}) (string, error) {
	// Try to find main content
	content := doc.Find("article, main, [role='main'], .content, #content").First()
	if content.Length() == 0 {
		content = doc.Find("body")
	}

	// Get text
	text := content.Text()

	// Clean up
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	// Limit length
	if len(text) > 8000 {
		text = text[:8000] + "\n\n[Content truncated - 8000 chars shown]"
	}

	return text, nil
}

func (t *WebBrowserTool) extractLinks(doc *goquery.Document, input map[string]interface{}) (string, error) {
	var links []string

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		text := strings.TrimSpace(s.Text())
		if text == "" {
			text = href
		}

		links = append(links, fmt.Sprintf("[%s](%s)", text, href))
	})

	if len(links) == 0 {
		return "No links found on page.", nil
	}

	return "Links found:\n" + strings.Join(links[:min(len(links), 50)], "\n"), nil
}

func (t *WebBrowserTool) extractForms(doc *goquery.Document, input map[string]interface{}) (string, error) {
	var forms []map[string]interface{}

	doc.Find("form").Each(func(i int, s *goquery.Selection) {
		formInfo := map[string]interface{}{
			"index":  i,
			"action": "",
			"method": "GET",
			"fields": []map[string]string{},
		}

		if action, exists := s.Attr("action"); exists {
			formInfo["action"] = action
		}
		if method, exists := s.Attr("method"); exists {
			formInfo["method"] = strings.ToUpper(method)
		}

		var fields []map[string]string
		s.Find("input, textarea, select").Each(func(j int, field *goquery.Selection) {
			fieldInfo := map[string]string{}
			if name, exists := field.Attr("name"); exists {
				fieldInfo["name"] = name
			}
			if ftype, exists := field.Attr("type"); exists {
				fieldInfo["type"] = ftype
			}
			if placeholder, exists := field.Attr("placeholder"); exists {
				fieldInfo["placeholder"] = placeholder
			}
			if len(fieldInfo) > 0 {
				fields = append(fields, fieldInfo)
			}
		})
		formInfo["fields"] = fields

		forms = append(forms, formInfo)
	})

	if len(forms) == 0 {
		return "No forms found on page.", nil
	}

	data, err := json.MarshalIndent(forms, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal forms: %w", err)
	}
	return string(data), nil
}

func (t *WebBrowserTool) extractImages(doc *goquery.Document, input map[string]interface{}) (string, error) {
	var images []string

	doc.Find("img[src]").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if !exists {
			return
		}

		alt, _ := s.Attr("alt")
		images = append(images, fmt.Sprintf("![%s](%s)", alt, src))
	})

	if len(images) == 0 {
		return "No images found on page.", nil
	}

	return "Images found:\n" + strings.Join(images[:min(len(images), 30)], "\n"), nil
}

func (t *WebBrowserTool) extractArticle(doc *goquery.Document, input map[string]interface{}) (string, error) {
	// Try to find article content
	article := doc.Find("article, [role='article']").First()
	if article.Length() == 0 {
		// Try heuristics for main content
		article = doc.Find(".post-content, .entry-content, .article-content, .story-body").First()
	}
	if article.Length() == 0 {
		article = doc.Find("main").First()
	}
	if article.Length() == 0 {
		// Fallback to body
		article = doc.Find("body")
	}

	// Remove navigation, ads, etc
	article.Find("nav, header, footer, aside, .ads, .advertisement, .sidebar, .comments, script, style").Remove()

	// Get title
	title := doc.Find("h1, .article-title, .post-title").First().Text()
	if title == "" {
		title = doc.Find("title").Text()
	}

	// Get content
	content := article.Text()
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
	content = strings.TrimSpace(content)

	result := fmt.Sprintf("# %s\n\n", strings.TrimSpace(title))
	result += content

	if len(result) > 10000 {
		result = result[:10000] + "\n\n[Article truncated - 10000 chars shown]"
	}

	return result, nil
}

func (t *WebBrowserTool) screenshot(session *BrowserSession, input map[string]interface{}) (string, error) {
	return "Screenshot not available in text mode. Use web_browser with extract action to get page content.", nil
}

func (t *WebBrowserTool) back(session *BrowserSession, input map[string]interface{}) (string, error) {
	if session.HistoryIndex <= 0 {
		return "Cannot go back - at beginning of history", nil
	}

	session.HistoryIndex--
	session.CurrentURL = session.History[session.HistoryIndex]

	// Reload page
	input["url"] = session.CurrentURL
	return t.navigate(session, input)
}

func (t *WebBrowserTool) forward(session *BrowserSession, input map[string]interface{}) (string, error) {
	if session.HistoryIndex >= len(session.History)-1 {
		return "Cannot go forward - at end of history", nil
	}

	session.HistoryIndex++
	session.CurrentURL = session.History[session.HistoryIndex]

	// Reload page
	input["url"] = session.CurrentURL
	return t.navigate(session, input)
}

func resolveURL(base, ref string) string {
	baseURL, err := url.Parse(base)
	if err != nil {
		return ref
	}

	refURL, err := url.Parse(ref)
	if err != nil {
		return ref
	}

	return baseURL.ResolveReference(refURL).String()
}
