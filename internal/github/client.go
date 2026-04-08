package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	githubAPIBaseURL = "https://api.github.com"
	tokenFileName    = "github_token"
)

// Client represents a GitHub API client
type Client struct {
	Token      string
	HTTPClient *http.Client
}

// NewClient creates a new GitHub client
func NewClient() *Client {
	token := loadToken()
	return &Client{
		Token: token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// IsAuthenticated checks if client has a token
func (c *Client) IsAuthenticated() bool {
	return c.Token != ""
}

// Authenticate sets the token and saves it
func (c *Client) Authenticate(token string) error {
	c.Token = token
	return saveToken(token)
}

// Logout removes the stored token
func (c *Client) Logout() error {
	c.Token = ""
	return deleteToken()
}

// loadToken loads the GitHub token from storage
func loadToken() string {
	tokenPath := getTokenPath()
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		// Try environment variable
		return os.Getenv("GITHUB_TOKEN")
	}
	return string(data)
}

// saveToken saves the GitHub token
func saveToken(token string) error {
	tokenPath := getTokenPath()
	dir := filepath.Dir(tokenPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return os.WriteFile(tokenPath, []byte(token), 0600)
}

// deleteToken removes the stored token
func deleteToken() error {
	tokenPath := getTokenPath()
	return os.Remove(tokenPath)
}

// getTokenPath returns the path to the token file
func getTokenPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".letsGo", tokenFileName)
}

// doRequest makes an authenticated HTTP request
func (c *Client) doRequest(method, url string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Claude-Code-Go/0.1.0")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add authentication
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	return c.HTTPClient.Do(req)
}

// GetAuthenticatedUser gets the current user
func (c *Client) GetAuthenticatedUser() (*User, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	resp, err := c.doRequest("GET", githubAPIBaseURL+"/user", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// ListRepos lists repositories for the authenticated user
func (c *Client) ListRepos() ([]Repo, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	resp, err := c.doRequest("GET", githubAPIBaseURL+"/user/repos?sort=updated&per_page=100", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var repos []Repo
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}

// GetRepo gets a specific repository
func (c *Client) GetRepo(owner, repo string) (*Repo, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", githubAPIBaseURL, owner, repo)

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var repository Repo
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, err
	}

	return &repository, nil
}

// ListIssues lists issues for a repository
func (c *Client) ListIssues(owner, repo string, state string) ([]Issue, error) {
	if state == "" {
		state = "open"
	}

	url := fmt.Sprintf("%s/repos/%s/%s/issues?state=%s", githubAPIBaseURL, owner, repo, state)

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var issues []Issue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	return issues, nil
}

// CreateIssue creates a new issue
func (c *Client) CreateIssue(owner, repo, title, body string) (*Issue, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	url := fmt.Sprintf("%s/repos/%s/%s/issues", githubAPIBaseURL, owner, repo)

	payload := map[string]string{
		"title": title,
		"body":  body,
	}

	resp, err := c.doRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	return &issue, nil
}

// ListPullRequests lists pull requests for a repository
func (c *Client) ListPullRequests(owner, repo, state string) ([]PullRequest, error) {
	if state == "" {
		state = "open"
	}

	url := fmt.Sprintf("%s/repos/%s/%s/pulls?state=%s", githubAPIBaseURL, owner, repo, state)

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var prs []PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, err
	}

	return prs, nil
}

// CreatePullRequest creates a new pull request
func (c *Client) CreatePullRequest(owner, repo, title, head, base, body string) (*PullRequest, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	url := fmt.Sprintf("%s/repos/%s/%s/pulls", githubAPIBaseURL, owner, repo)

	payload := map[string]string{
		"title": title,
		"head":  head,
		"base":  base,
		"body":  body,
	}

	resp, err := c.doRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}

	return &pr, nil
}

// GetFileContent gets the content of a file
func (c *Client) GetFileContent(owner, repo, path, ref string) (*FileContent, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", githubAPIBaseURL, owner, repo, path)
	if ref != "" {
		url += "?ref=" + ref
	}

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var content FileContent
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return nil, err
	}

	return &content, nil
}

// SearchCode searches code across GitHub
func (c *Client) SearchCode(query string) (*CodeSearchResult, error) {
	url := fmt.Sprintf("%s/search/code?q=%s", githubAPIBaseURL, query)

	resp, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var result CodeSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Data structures

// User represents a GitHub user
type User struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	AvatarURL string `json:"avatar_url"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Bio       string `json:"bio"`
	HTMLURL   string `json:"html_url"`
}

// Repo represents a GitHub repository
type Repo struct {
	ID            int       `json:"id"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Description   string    `json:"description"`
	Private       bool      `json:"private"`
	HTMLURL       string    `json:"html_url"`
	CloneURL      string    `json:"clone_url"`
	SSHURL        string    `json:"ssh_url"`
	Stars         int       `json:"stargazers_count"`
	Forks         int       `json:"forks_count"`
	OpenIssues    int       `json:"open_issues_count"`
	Language      string    `json:"language"`
	DefaultBranch string    `json:"default_branch"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Issue represents a GitHub issue
type Issue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	HTMLURL   string    `json:"html_url"`
	User      User      `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Labels    []Label   `json:"labels"`
}

// Label represents a GitHub label
type Label struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	HTMLURL   string    `json:"html_url"`
	User      User      `json:"user"`
	Head      Branch    `json:"head"`
	Base      Branch    `json:"base"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Branch represents a branch reference
type Branch struct {
	Ref  string `json:"ref"`
	SHA  string `json:"sha"`
	Repo Repo   `json:"repo"`
}

// FileContent represents file content
type FileContent struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	SHA     string `json:"sha"`
	Size    int    `json:"size"`
	Type    string `json:"type"`
	Content string `json:"content"`
	HTMLURL string `json:"html_url"`
}

// CodeSearchResult represents code search results
type CodeSearchResult struct {
	TotalCount int          `json:"total_count"`
	Items      []CodeResult `json:"items"`
}

// CodeResult represents a code search result item
type CodeResult struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	HTMLURL     string `json:"html_url"`
	Repository  Repo   `json:"repository"`
	TextMatches []struct {
		Fragment string `json:"fragment"`
		Matches  []struct {
			Text string `json:"text"`
		} `json:"matches"`
	} `json:"text_matches"`
}
