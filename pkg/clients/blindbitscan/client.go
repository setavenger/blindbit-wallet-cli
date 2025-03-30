package client

import (
	"net/http"
	"time"

	"github.com/setavenger/blindbit-scan/pkg/wallet" // for OwnedUTXO
)

// basicAuthTransport is a custom RoundTripper that adds a Basic Auth header to every request.
type basicAuthTransport struct {
	username string
	password string
	rt       http.RoundTripper
}

// RoundTrip adds the Authorization header and delegates to the underlying RoundTripper.
func (bat *basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(bat.username, bat.password)
	return bat.rt.RoundTrip(req)
}

// Client wraps the HTTP client and API base URL.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient returns a new API client. If username and password are non-empty,
// the client will send a Basic Auth header with every request.
func NewClient(baseURL, username, password string) *Client {
	// Use the default transport, or wrap it with basicAuthTransport if auth is provided.
	var transport http.RoundTripper = http.DefaultTransport
	if username != "" && password != "" {
		transport = &basicAuthTransport{
			username: username,
			password: password,
			rt:       http.DefaultTransport,
		}
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		},
	}
}

// --- Response/Request Structs ---

// heightResponse mirrors the JSON response from GET /height.
type heightResponse struct {
	Height uint64 `json:"height"`
}

// utxosResponse represents a slice of UTXOs returned by GET /utxos.
type utxosResponse []*wallet.OwnedUTXO

// addressResponse mirrors the JSON response from GET /address and PUT /silentpaymentkeys.
type addressResponse struct {
	Address string `json:"address"`
}

// RescanReq is used for POST /rescan.
type RescanReq struct {
	Height uint64 `json:"height"`
}

// SetupReq is used for PUT /silentpaymentkeys.
type SetupReq struct {
	ScanSecret  string `json:"secret_sec"`
	SpendPublic string `json:"spend_pub"`
	BirthHeight uint   `json:"birth_height"`
}
