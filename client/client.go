package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	// DefaultBaseURL is the default base URL for the OpenAI API
	DefaultBaseURL = "https://api.openai.com/v1"
	// DefaultUserAgent is the default user agent for the OpenAI API client
	DefaultUserAgent = "openai-responses-api-go/1.0.0"
	// DefaultTimeout is the default timeout for API requests
	DefaultTimeout = 30 * time.Second
)

// Client is the client for the OpenAI Responses API
type Client struct {
	// BaseURL is the base URL for API requests
	BaseURL string
	// APIKey is the API key for authentication
	APIKey string
	// HTTPClient is the HTTP client used to make API requests
	HTTPClient *http.Client
	// UserAgent is the user agent for API requests
	UserAgent string
	// Organization is the organization ID for API requests
	Organization string
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithBaseURL sets the base URL for the client
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) {
		c.BaseURL = baseURL
	}
}

// WithAPIKey sets the API key for the client
func WithAPIKey(apiKey string) ClientOption {
	return func(c *Client) {
		c.APIKey = apiKey
	}
}

// WithHTTPClient sets the HTTP client for the client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.HTTPClient = httpClient
	}
}

// WithUserAgent sets the user agent for the client
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.UserAgent = userAgent
	}
}

// WithOrganization sets the organization ID for the client
func WithOrganization(organization string) ClientOption {
	return func(c *Client) {
		c.Organization = organization
	}
}

// NewClient creates a new OpenAI Responses API client
func NewClient(options ...ClientOption) *Client {
	client := &Client{
		BaseURL:    DefaultBaseURL,
		UserAgent:  DefaultUserAgent,
		HTTPClient: &http.Client{Timeout: DefaultTimeout},
	}

	// Apply options
	for _, option := range options {
		option(client)
	}

	// If API key is not set, try to get it from environment variable
	if client.APIKey == "" {
		client.APIKey = os.Getenv("OPENAI_API_KEY")
	}

	return client
}

// APIError represents an error returned by the OpenAI API
type APIError struct {
	Code       *string `json:"code,omitempty"`
	Message    string  `json:"message"`
	Param      *string `json:"param,omitempty"`
	Type       string  `json:"type"`
	StatusCode int     `json:"-"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Code != nil {
		return fmt.Sprintf("OpenAI API error: code=%s message=%s param=%v type=%s status_code=%d",
			*e.Code, e.Message, e.Param, e.Type, e.StatusCode)
	}
	return fmt.Sprintf("OpenAI API error: message=%s param=%v type=%s status_code=%d",
		e.Message, e.Param, e.Type, e.StatusCode)
}

// ErrorResponse represents the error response from the OpenAI API
type ErrorResponse struct {
	Error *APIError `json:"error,omitempty"`
}

// request makes an HTTP request to the OpenAI API
func (c *Client) request(ctx context.Context, method, path string, body interface{}, v interface{}) error {
	// Construct the URL
	u, err := url.Parse(c.BaseURL + path)
	if err != nil {
		return err
	}

	// Create the request body
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, method, u.String(), reqBody)
	if err != nil {
		return err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	if c.Organization != "" {
		req.Header.Set("OpenAI-Organization", c.Organization)
	}

	// Make the request
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check for errors
	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("error decoding error response: %w", err)
		}
		if errResp.Error != nil {
			errResp.Error.StatusCode = resp.StatusCode
			return errResp.Error
		}
		return fmt.Errorf("unknown error, status code: %d", resp.StatusCode)
	}

	// Decode the response
	if v != nil {
		if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
			return err
		}
	}

	return nil
}

// get makes a GET request to the OpenAI API
func (c *Client) get(ctx context.Context, path string, v interface{}) error {
	return c.request(ctx, http.MethodGet, path, nil, v)
}

// post makes a POST request to the OpenAI API
func (c *Client) post(ctx context.Context, path string, body interface{}, v interface{}) error {
	return c.request(ctx, http.MethodPost, path, body, v)
}

// delete makes a DELETE request to the OpenAI API
func (c *Client) delete(ctx context.Context, path string, v interface{}) error {
	return c.request(ctx, http.MethodDelete, path, nil, v)
}