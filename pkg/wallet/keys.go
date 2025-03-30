package wallet

import (
	"crypto/sha256"
	"fmt"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/setavenger/go-bip352"
	"github.com/tyler-smith/go-bip39"
)

var (
	// Network parameters for different networks
	networkParams = map[Network]*chaincfg.Params{
		NetworkMainnet: &chaincfg.MainNetParams,
		NetworkTestnet: &chaincfg.TestNet3Params,
		NetworkSignet:  &chaincfg.SigNetParams,
	}
)

// DeriveKeys derives scan and spend secrets from a mnemonic
func DeriveKeys(mnemonic string) (scanSecret, spendSecret []byte, err error) {
	// Validate mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, nil, fmt.Errorf("invalid mnemonic")
	}

	// Generate seed from mnemonic
	seed := bip39.NewSeed(mnemonic, "")

	// Derive scan secret (first 32 bytes of SHA256(seed))
	scanHash := sha256.Sum256(seed)
	scanSecret = scanHash[:]

	// Derive spend secret (next 32 bytes of SHA256(seed))
	spendHash := sha256.Sum256(scanHash[:])
	spendSecret = spendHash[:]

	return scanSecret, spendSecret, nil
}

// GenerateAddress generates a silent payment address from the scan secret
func GenerateAddress(scanSecret []byte, network Network) (string, error) {
	params, ok := networkParams[network]
	if !ok {
		return "", fmt.Errorf("unsupported network: %s", network)
	}

	// Generate scan public key
	scanPrivKey, _ := btcec.PrivKeyFromBytes(scanSecret)
	scanPubKey := scanPrivKey.PubKey()

	// Generate silent payment address
	tapKey := txscript.ComputeTaprootKeyNoScript(scanPubKey)
	addr, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(tapKey), params)
	if err != nil {
		return "", fmt.Errorf("failed to create address: %w", err)
	}

	return addr.String(), nil
}

// GenerateLabel generates a labeled address
func GenerateLabel(scanSecret, spendSecret []byte, m uint32, network Network) (*bip352.Label, error) {
	params, ok := networkParams[network]
	if !ok {
		return nil, fmt.Errorf("unsupported network: %s", network)
	}

	// Generate scan and spend public keys
	scanPrivKey, _ := btcec.PrivKeyFromBytes(scanSecret)
	spendPrivKey, _ := btcec.PrivKeyFromBytes(spendSecret)

	// Generate tweak
	tweak := generateTweak(scanPrivKey.PubKey(), spendPrivKey.PubKey(), m)

	// Generate labeled public key
	labeledPubKey, err := generateLabeledPubKey(spendPrivKey.PubKey(), tweak)
	if err != nil {
		return nil, fmt.Errorf("failed to generate labeled public key: %w", err)
	}

	// Generate silent payment address
	tapKey := txscript.ComputeTaprootKeyNoScript(labeledPubKey)
	addr, err := btcutil.NewAddressTaproot(schnorr.SerializePubKey(tapKey), params)
	if err != nil {
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	// Convert the byte slices to fixed-size arrays
	var pubKey [33]byte
	var tweakBytes [32]byte
	copy(pubKey[:], schnorr.SerializePubKey(labeledPubKey))
	copy(tweakBytes[:], tweak)

	return &bip352.Label{
		PubKey:  pubKey,
		Tweak:   tweakBytes,
		Address: addr.String(),
		M:       m,
	}, nil
}

// generateTweak generates the tweak for a labeled address
func generateTweak(scanPubKey, spendPubKey *btcec.PublicKey, m uint32) []byte {
	// Concatenate public keys and m value
	data := append(scanPubKey.SerializeCompressed(), spendPubKey.SerializeCompressed()...)
	data = append(data, byte(m))

	// Hash the data
	hash := sha256.Sum256(data)
	return hash[:]
}

// generateLabeledPubKey generates the labeled public key
func generateLabeledPubKey(spendPubKey *btcec.PublicKey, tweak []byte) (*btcec.PublicKey, error) {
	// Convert tweak to scalar
	var tweakScalar btcec.ModNScalar
	if overflow := tweakScalar.SetByteSlice(tweak); overflow {
		return nil, fmt.Errorf("tweak value is too large")
	}

	// Create a point from the tweak scalar by multiplying with generator
	var tweakPoint btcec.JacobianPoint
	btcec.ScalarBaseMultNonConst(&tweakScalar, &tweakPoint)

	// Convert spend public key to Jacobian point
	var spendPoint btcec.JacobianPoint
	spendPubKey.AsJacobian(&spendPoint)

	// Add the points
	var resultPoint btcec.JacobianPoint
	btcec.AddNonConst(&spendPoint, &tweakPoint, &resultPoint)

	// Convert back to affine coordinates
	resultPoint.ToAffine()

	// Create new public key from the result point
	return btcec.NewPublicKey(&resultPoint.X, &resultPoint.Y), nil
}

// DerivePublicKey derives a public key from the spend secret and tweak
func DerivePublicKey(spendSecret []byte, tweak [32]byte) (*btcec.PublicKey, error) {
	// Convert spend secret to fixed-size array
	var spendSecretArr [32]byte
	copy(spendSecretArr[:], spendSecret)

	// Add the private keys (spend secret and tweak)
	mergedSecret := bip352.AddPrivateKeys(spendSecretArr, tweak)

	_, mergedPubKey := btcec.PrivKeyFromBytes(mergedSecret[:])

	return mergedPubKey, nil
}

// GetScanOnlyParams returns the scan-only parameters for the wallet
func (w *Wallet) GetScanOnlyParams() *ScanOnlyParams {
	return &ScanOnlyParams{
		ScanSecret: w.ScanSecret,
		Network:    w.Network,
	}
}
