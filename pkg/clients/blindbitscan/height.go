package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetCurrentHeight calls the GET /height endpoint.
func (c *Client) GetCurrentHeight() (uint64, error) {
	url := fmt.Sprintf("%s/height", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var hr heightResponse
	if err := json.NewDecoder(resp.Body).Decode(&hr); err != nil {
		return 0, err
	}
	return hr.Height, nil
}
