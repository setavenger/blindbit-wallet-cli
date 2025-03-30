package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/setavenger/blindbit-wallet-cli/pkg/wallet"
)

const (
	walletFileName = "wallet.json"
)

// Storage handles encrypted wallet data storage
type Storage struct {
	baseDir string
}

// New creates a new storage instance
func New(baseDir string) (*Storage, error) {
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}
	return &Storage{baseDir: baseDir}, nil
}

// SaveWallet saves the wallet data with encryption
func (s *Storage) SaveWallet(data *wallet.WalletData, password string) error {
	// Marshal wallet data
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal wallet data: %w", err)
	}

	// Generate encryption key from password
	key := deriveKey(password)

	// Encrypt data
	encrypted, err := encrypt(key, jsonData)
	if err != nil {
		return fmt.Errorf("failed to encrypt wallet data: %w", err)
	}

	// Save to file
	walletPath := filepath.Join(s.baseDir, walletFileName)
	if err := os.WriteFile(walletPath, encrypted, 0600); err != nil {
		return fmt.Errorf("failed to write wallet file: %w", err)
	}

	return nil
}

// LoadWallet loads and decrypts wallet data
func (s *Storage) LoadWallet(password string) (*wallet.WalletData, error) {
	walletPath := filepath.Join(s.baseDir, walletFileName)
	encrypted, err := os.ReadFile(walletPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wallet file: %w", err)
	}

	// Generate decryption key from password
	key := deriveKey(password)

	// Decrypt data
	decrypted, err := decrypt(key, encrypted)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt wallet data: %w", err)
	}

	// Unmarshal wallet data
	var data wallet.WalletData
	if err := json.Unmarshal(decrypted, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet data: %w", err)
	}

	return &data, nil
}

// deriveKey generates an encryption key from a password
func deriveKey(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

// encrypt encrypts data using AES-256-GCM
func encrypt(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-256-GCM
func decrypt(key, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(data) < gcm.NonceSize() {
		return nil, fmt.Errorf("malformed ciphertext")
	}

	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// WalletExists checks if a wallet file exists
func (s *Storage) WalletExists() bool {
	walletPath := filepath.Join(s.baseDir, walletFileName)
	_, err := os.Stat(walletPath)
	return err == nil
}
