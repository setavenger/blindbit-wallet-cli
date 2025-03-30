package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// GetAddress calls the GET /address endpoint.
func (c *Client) GetAddress() (string, error) {
	url := fmt.Sprintf("%s/address", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var ar addressResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return "", err
	}
	return ar.Address, nil
}
