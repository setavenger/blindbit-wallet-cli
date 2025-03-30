package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/setavenger/blindbit-scan/pkg/wallet"
)

// GetUtxos calls the GET /utxos endpoint.
func (c *Client) GetUtxos() ([]*wallet.OwnedUTXO, error) {
	url := fmt.Sprintf("%s/utxos", c.baseURL)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var utxos utxosResponse
	if err := json.NewDecoder(resp.Body).Decode(&utxos); err != nil {
		return nil, err
	}
	return utxos, nil
}
