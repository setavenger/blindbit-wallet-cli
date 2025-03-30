package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// PutSilentPaymentKeys calls the PUT /silentpaymentkeys endpoint.
func (c *Client) PutSilentPaymentKeys(req SetupReq) (string, error) {
	url := fmt.Sprintf("%s/silentpaymentkeys", c.baseURL)
	data, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var ar addressResponse
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return "", err
	}
	return ar.Address, nil
}
