package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PostRescan calls the POST /rescan endpoint.
func (c *Client) PostRescan(height uint64) (uint64, error) {
	url := fmt.Sprintf("%s/rescan", c.baseURL)
	reqBody := RescanReq{Height: height}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return 0, err
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Read the response and unmarshal it
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var res RescanReq
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, err
	}
	return res.Height, nil
}
