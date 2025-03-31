package client

import (
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/setavenger/blindbit-wallet-cli/internal/config"
	"golang.org/x/net/proxy"
)

// TorClient wraps the Tor client functionality
type TorClient struct {
	proxyURL string
}

// NewTorClient creates a new Tor client
func NewTorClient(cfg *config.Config) (*TorClient, error) {
	if !cfg.UseTor {
		return nil, fmt.Errorf("tor is not enabled")
	}

	// Use the system's Tor SOCKS proxy
	proxyURL := fmt.Sprintf("socks5://%s:%d", cfg.TorHost, cfg.TorPort)
	return &TorClient{
		proxyURL: proxyURL,
	}, nil
}

// CreateHTTPClient creates an HTTP client that uses Tor
func (c *TorClient) CreateHTTPClient() *http.Client {
	proxyURL, err := url.Parse(c.proxyURL)
	if err != nil {
		// This should never happen as we construct the URL ourselves
		panic(fmt.Sprintf("invalid proxy URL: %v", err))
	}

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}
}

// Dial creates a new connection through Tor
func (c *TorClient) Dial(network, addr string) (net.Conn, error) {
	proxyURL, err := url.Parse(c.proxyURL)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}

	dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy dialer: %w", err)
	}

	return dialer.Dial(network, addr)
}

// Close closes the Tor client
func (c *TorClient) Close() error {
	// Nothing to close as we're using the system's Tor daemon
	return nil
}
