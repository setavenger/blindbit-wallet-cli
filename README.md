# BlindBit Wallet CLI

A command-line interface for managing Bitcoin Silent Payment (BIP 352) wallets.

## Features

- Create and import Silent Payment wallets
- Generate Silent Payment addresses
- Send Bitcoin to regular and Silent Payment addresses
- Sync with BlindBit Scan daemon
- View UTXOs and transaction history
- Support for Tor connections to .onion addresses

## Installation

```bash
go install github.com/setavenger/blindbit-wallet-cli@latest
```

## Configuration

Copy the example configuration file and customize it for your needs:

```bash
cp blindbit.toml.example ~/.blindbit-wallet/blindbit.toml
```

### Tor Support

To connect to .onion addresses, enable Tor in your configuration:

```toml
use_tor = true
tor_host = "localhost"  # Tor SOCKS proxy host
tor_port = 9050        # Tor SOCKS proxy port
```

Make sure you have Tor running on your system. On macOS, you can install it with:

```bash
brew install tor
```

Then start the Tor service:

```bash
brew services start tor
```

## Usage

### Create a new wallet

```bash
blindbit-wallet-cli wallet new
```

### Import an existing wallet

```bash
blindbit-wallet-cli wallet import
```

### Generate a Silent Payment address

```bash
blindbit-wallet-cli wallet address
```

### Send Bitcoin

```bash
blindbit-wallet-cli wallet send <address> <amount> [--fee-rate <rate>] [--label <label>]
```

### View UTXOs

```bash
blindbit-wallet-cli wallet utxos
```

### Sync with BlindBit Scan

```bash
blindbit-wallet-cli wallet sync
```

## License

MIT 