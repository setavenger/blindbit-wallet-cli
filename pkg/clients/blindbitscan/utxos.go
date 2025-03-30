package client

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/setavenger/blindbit-scan/pkg/wallet"
	"github.com/setavenger/go-bip352"
)

// GetUTXOs calls the GET /utxos endpoint.
func (c *Client) GetUTXOs(ctx context.Context) ([]*wallet.OwnedUTXO, error) {
	url := fmt.Sprintf("%s/utxos", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var utxos utxosResponse
	if err := json.NewDecoder(resp.Body).Decode(&utxos); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert utxosResponse to []*wallet.OwnedUTXO
	result := make([]*wallet.OwnedUTXO, len(utxos))
	for i, u := range utxos {
		var label *bip352.Label
		if u.Label != nil {
			label, err = wallet.ConvertLabelJSONToLabel(*u.Label)
			if err != nil {
				return nil, fmt.Errorf("failed to convert label: %w", err)
			}
		}

		txid, err := hex.DecodeString(u.Txid)
		if err != nil {
			return nil, fmt.Errorf("failed to decode txid: %w", err)
		}

		pubKey, err := hex.DecodeString(u.PubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decode pubkey: %w", err)
		}

		privKeyTweak, err := hex.DecodeString(u.PrivKeyTweak)
		if err != nil {
			return nil, fmt.Errorf("failed to decode priv_key_tweak: %w", err)
		}

		result[i] = &wallet.OwnedUTXO{
			Txid:         bip352.ConvertToFixedLength32(txid),
			Vout:         u.Vout,
			Amount:       u.Amount,
			PrivKeyTweak: bip352.ConvertToFixedLength32(privKeyTweak),
			PubKey:       bip352.ConvertToFixedLength32(pubKey),
			Timestamp:    u.Timestamp,
			State:        u.State,
			Label:        label,
		}
	}

	return result, nil
}
