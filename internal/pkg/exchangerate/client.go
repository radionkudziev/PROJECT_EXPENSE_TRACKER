package exchangerate

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

type Response struct {
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
}

func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *Client) GetRate(from, to string) (float64, error) {
	url := fmt.Sprintf("%s%s", c.baseURL, from)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	if c.apiKey != "" {
		req.Header.Set("apikey", c.apiKey)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to call external API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("external API returned status %d", resp.StatusCode)
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	rate, ok := result.Rates[to]
	if !ok {
		return 0, fmt.Errorf("rate for %s not found", to)
	}
	return rate, nil
}
