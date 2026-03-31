package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURL = "https://api.mainlayer.xyz"
	defaultTimeout = 30 * time.Second
	userAgent      = "terraform-provider-mainlayer/1.0"
)

// Client is the Mainlayer API HTTP client.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClient constructs a new Mainlayer API client.
func NewClient(apiKey, baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// APIError represents an error returned by the Mainlayer API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("mainlayer API error (status %d): %s", e.StatusCode, e.Body)
}

// Resource represents a Mainlayer resource.
type Resource struct {
	ID          string  `json:"id,omitempty"`
	Slug        string  `json:"slug"`
	Type        string  `json:"type"`
	PriceUSDC   float64 `json:"price_usdc"`
	FeeModel    string  `json:"fee_model"`
	Description string  `json:"description,omitempty"`
	CallbackURL string  `json:"callback_url,omitempty"`
	CreatedAt   string  `json:"created_at,omitempty"`
	UpdatedAt   string  `json:"updated_at,omitempty"`
}

// Plan represents a Mainlayer plan (subscription tier for a resource).
type Plan struct {
	ID          string  `json:"id,omitempty"`
	ResourceID  string  `json:"resource_id"`
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	PriceUSDC   float64 `json:"price_usdc"`
	CallLimit   int64   `json:"call_limit,omitempty"`
	Period      string  `json:"period,omitempty"`
	CreatedAt   string  `json:"created_at,omitempty"`
	UpdatedAt   string  `json:"updated_at,omitempty"`
}

// ListResourcesResponse holds the paginated list of resources.
type ListResourcesResponse struct {
	Resources []Resource `json:"resources"`
	Total     int        `json:"total"`
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	return c.httpClient.Do(req)
}

func readAndClose(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func checkStatus(resp *http.Response, data []byte) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return &APIError{StatusCode: resp.StatusCode, Body: string(data)}
}

// --- Resource CRUD ---

// CreateResource creates a new resource via the API.
func (c *Client) CreateResource(ctx context.Context, r *Resource) (*Resource, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/v1/resources", r)
	if err != nil {
		return nil, err
	}
	data, err := readAndClose(resp)
	if err != nil {
		return nil, err
	}
	if err := checkStatus(resp, data); err != nil {
		return nil, err
	}
	var created Resource
	if err := json.Unmarshal(data, &created); err != nil {
		return nil, fmt.Errorf("decoding create resource response: %w", err)
	}
	return &created, nil
}

// GetResource retrieves a resource by ID.
func (c *Client) GetResource(ctx context.Context, id string) (*Resource, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/v1/resources/"+id, nil)
	if err != nil {
		return nil, err
	}
	data, err := readAndClose(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if err := checkStatus(resp, data); err != nil {
		return nil, err
	}
	var resource Resource
	if err := json.Unmarshal(data, &resource); err != nil {
		return nil, fmt.Errorf("decoding get resource response: %w", err)
	}
	return &resource, nil
}

// UpdateResource updates a resource by ID.
func (c *Client) UpdateResource(ctx context.Context, id string, r *Resource) (*Resource, error) {
	resp, err := c.doRequest(ctx, http.MethodPut, "/v1/resources/"+id, r)
	if err != nil {
		return nil, err
	}
	data, err := readAndClose(resp)
	if err != nil {
		return nil, err
	}
	if err := checkStatus(resp, data); err != nil {
		return nil, err
	}
	var updated Resource
	if err := json.Unmarshal(data, &updated); err != nil {
		return nil, fmt.Errorf("decoding update resource response: %w", err)
	}
	return &updated, nil
}

// DeleteResource deletes a resource by ID.
func (c *Client) DeleteResource(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, "/v1/resources/"+id, nil)
	if err != nil {
		return err
	}
	data, err := readAndClose(resp)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	return checkStatus(resp, data)
}

// ListResources returns all resources the API key can access.
func (c *Client) ListResources(ctx context.Context) ([]Resource, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/v1/resources", nil)
	if err != nil {
		return nil, err
	}
	data, err := readAndClose(resp)
	if err != nil {
		return nil, err
	}
	if err := checkStatus(resp, data); err != nil {
		return nil, err
	}
	var listResp ListResourcesResponse
	if err := json.Unmarshal(data, &listResp); err != nil {
		// Fallback: maybe the API returns a plain array.
		var resources []Resource
		if err2 := json.Unmarshal(data, &resources); err2 != nil {
			return nil, fmt.Errorf("decoding list resources response: %w", err)
		}
		return resources, nil
	}
	return listResp.Resources, nil
}

// --- Plan CRUD ---

// CreatePlan creates a plan for a resource.
func (c *Client) CreatePlan(ctx context.Context, p *Plan) (*Plan, error) {
	path := fmt.Sprintf("/v1/resources/%s/plans", p.ResourceID)
	resp, err := c.doRequest(ctx, http.MethodPost, path, p)
	if err != nil {
		return nil, err
	}
	data, err := readAndClose(resp)
	if err != nil {
		return nil, err
	}
	if err := checkStatus(resp, data); err != nil {
		return nil, err
	}
	var created Plan
	if err := json.Unmarshal(data, &created); err != nil {
		return nil, fmt.Errorf("decoding create plan response: %w", err)
	}
	return &created, nil
}

// GetPlan retrieves a plan by resource ID and plan ID.
func (c *Client) GetPlan(ctx context.Context, resourceID, planID string) (*Plan, error) {
	path := fmt.Sprintf("/v1/resources/%s/plans/%s", resourceID, planID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	data, err := readAndClose(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if err := checkStatus(resp, data); err != nil {
		return nil, err
	}
	var plan Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("decoding get plan response: %w", err)
	}
	return &plan, nil
}

// UpdatePlan updates a plan by resource ID and plan ID.
func (c *Client) UpdatePlan(ctx context.Context, resourceID, planID string, p *Plan) (*Plan, error) {
	path := fmt.Sprintf("/v1/resources/%s/plans/%s", resourceID, planID)
	resp, err := c.doRequest(ctx, http.MethodPut, path, p)
	if err != nil {
		return nil, err
	}
	data, err := readAndClose(resp)
	if err != nil {
		return nil, err
	}
	if err := checkStatus(resp, data); err != nil {
		return nil, err
	}
	var updated Plan
	if err := json.Unmarshal(data, &updated); err != nil {
		return nil, fmt.Errorf("decoding update plan response: %w", err)
	}
	return &updated, nil
}

// DeletePlan deletes a plan by resource ID and plan ID.
func (c *Client) DeletePlan(ctx context.Context, resourceID, planID string) error {
	path := fmt.Sprintf("/v1/resources/%s/plans/%s", resourceID, planID)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	data, err := readAndClose(resp)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	return checkStatus(resp, data)
}
