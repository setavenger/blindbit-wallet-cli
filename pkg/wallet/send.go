package wallet

import (
	"bytes"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/btcutil/txsort"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/mempool"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	scanwallet "github.com/setavenger/blindbit-scan/pkg/wallet"
	"github.com/setavenger/blindbit-wallet-cli/pkg/coinselector"
	"github.com/setavenger/go-bip352"
)

// Recipient represents a transaction recipient
type Recipient struct {
	Address string
	Amount  uint64
}

// SendToRecipients sends Bitcoin to the given recipients
func SendToRecipients(
	walletData *WalletData,
	recipients []Recipient,
	feeRate uint32,
) (
	[]byte,
	error,
) {
	// Convert recipients to coin selector format
	var selectorRecipients []*coinselector.Recipient
	for _, r := range recipients {
		selectorRecipients = append(selectorRecipients, &coinselector.Recipient{
			Address: r.Address,
			Amount:  r.Amount,
		})
	}

	// Convert UTXOs to coin selector format
	var utxos scanwallet.UtxoCollection
	for _, u := range walletData.UTXOs {
		if u.State != scanwallet.StateUnspent {
			continue
		}
		utxos = append(utxos, &u)
	}

	// Get chain parameters
	var chainParams *chaincfg.Params
	switch walletData.Wallet.Network {
	case NetworkMainnet:
		chainParams = &chaincfg.MainNetParams
	case NetworkTestnet:
		chainParams = &chaincfg.TestNet3Params
	case NetworkSignet:
		chainParams = &chaincfg.SigNetParams
	default:
		return nil, fmt.Errorf("unsupported network: %s", walletData.Wallet.Network)
	}

	// Send to recipients
	return walletData.Wallet.SendToRecipients(
		selectorRecipients,
		utxos,
		int64(feeRate),
		chainParams,
		546,   // Minimum change amount
		false, // Don't mark as spent
		false, // Don't use unconfirmed spent
	)
}

func (w Wallet) SendToRecipients(
	recipients []*coinselector.Recipient,
	utxos scanwallet.UtxoCollection,
	feeRate int64,
	chainParams *chaincfg.Params,
	minChangeAmount uint64,
	markSpent, useSpentUnconfirmed bool,
) (
	txBytes []byte,
	err error,
) {
	selector := coinselector.NewFeeRateCoinSelector(utxos, minChangeAmount, recipients)

	selectedUTXOs, changeAmount, err := selector.CoinSelect(uint32(feeRate))
	if err != nil {
		return nil, err
	}

	// vins is the final selection of coins, which can then be used to derive silentPayment Outputs
	var vins = make([]*bip352.Vin, len(selectedUTXOs))
	for i, utxo := range selectedUTXOs {
		vin := ConvertOwnedUTXOIntoVin(utxo)
		fullVinSecretKey := bip352.AddPrivateKeys(*vin.SecretKey, [32]byte(w.SpendSecret))
		vin.SecretKey = &fullVinSecretKey
		vins[i] = &vin
	}

	// now we need the difference between the inputs and outputs so that we can assign a value for change
	var sumAllInputs uint64
	for _, vin := range vins {
		sumAllInputs += vin.Amount
	}

	fmt.Println(w.ChangeAddress())

	if changeAmount > 0 {
		// change exists, and it should be greater than the MinChangeAmount
		recipients = append(recipients, &coinselector.Recipient{
			Address: w.ChangeAddress(),
			Amount:  changeAmount,
		})
	}

	// extract the ScriptPubKeys of the SP recipients with the selected txInputs
	recipients, err = ParseRecipients(recipients, vins, chainParams)
	if err != nil {
		return nil, err
	}

	err = sanityCheckRecipientsForSending(recipients)
	if err != nil {
		return nil, err
	}

	packet, err := CreateUnsignedPsbt(recipients, vins)
	if err != nil {
		return nil, err
	}

	err = SignPsbt(packet, vins)
	if err != nil {
		return nil, err
	}

	err = psbt.MaybeFinalizeAll(packet)
	if err != nil {
		panic(err) // todo remove panic
	}

	finalTx, err := psbt.Extract(packet)
	if err != nil {
		panic(err) // todo remove panic
	}

	var sumAllOutputs uint64
	for _, recipient := range recipients {
		sumAllOutputs += recipient.Amount
	}
	vSize := mempool.GetTxVirtualSize(btcutil.NewTx(finalTx))
	actualFee := sumAllInputs - sumAllOutputs
	actualFeeRate := float64(actualFee) / float64(vSize)

	errorTerm := 0.25 // todo make variable
	if actualFeeRate > float64(feeRate)+errorTerm {
		err = fmt.Errorf("actual fee rate deviates to strong from desired fee rate: %f > %d", actualFeeRate, feeRate)
		return nil, err
	}

	if actualFeeRate < float64(feeRate)-errorTerm {
		err = fmt.Errorf("actual fee rate deviates to strong from desired fee rate: %f < %d", actualFeeRate, feeRate)
		return nil, err
	}

	var buf bytes.Buffer
	err = finalTx.Serialize(&buf)
	if err != nil {
		return nil, err
	}

	// if markSpent {
	// 	var found int
	// 	// now that everything worked mark as spent if desired
	// 	for _, vin := range vins {
	// 		vinOutpoint, err := utils.SerialiseVinToOutpoint(*vin)
	// 		if err != nil {
	// 			logging.ErrorLogger.Println(err)
	// 			return nil, err
	// 		}
	// 		for _, utxo := range d.Wallet.UTXOs {
	// 			utxoOutpoint, err := utxo.SerialiseToOutpoint()
	// 			if err != nil {
	// 				logging.ErrorLogger.Println(err)
	// 				return nil, err
	// 			}
	// 			if bytes.Equal(vinOutpoint[:], utxoOutpoint[:]) {
	// 				utxo.State = src.StateUnconfirmedSpent
	// 				found++
	// 				logging.DebugLogger.Printf("Marked %x as spent\n", utxoOutpoint)
	// 			}
	// 		}
	// 	}
	// 	if found != len(vins) {
	// 		err = fmt.Errorf("we could not mark enough utxos as spent. marked %d, needed %d", found, len(vins))
	// 		return nil, err
	// 	}
	// }

	return buf.Bytes(), err
}

// Taken from blindbitd
//
// ParseRecipients
// Checks all recipients and adds the PkScript based on the given address.
// Silent Payment addresses are also parsed and the outputs will be computed based on the vins.
// For that reason this function has to be called after the final coinSelection is done.
// Otherwise, the SP outputs will NOT be found by the receiver.
// SP Recipients are always at the end.
// Hence, the tx must be sorted according to BIP 69 to avoid a specific signature of this wallet.
//
// NOTE: Existing PkScripts will NOT be overridden, those recipients will be skipped and returned as given
// todo keep original order in case that is relevant for any use case?
func ParseRecipients(
	recipients []*coinselector.Recipient,
	vins []*bip352.Vin,
	chainParams *chaincfg.Params,
) (
	[]*coinselector.Recipient,
	error,
) {
	var spRecipients []*bip352.Recipient

	// newRecipients tracks the modified group of recipients in order to avoid clashes
	var newRecipients []*coinselector.Recipient
	for _, recipient := range recipients {
		if recipient.PkScript != nil {
			// skip if a pkScript is already present (for what ever reason)
			newRecipients = append(newRecipients, recipient)
			continue
		}
		isSP := bip352.IsSilentPaymentAddress(recipient.Address)
		if !isSP {
			address, err := btcutil.DecodeAddress(recipient.Address, chainParams)
			if err != nil {
				log.Printf("Failed to decode address: %v", err)
				return nil, err
			}
			scriptPubKey, err := txscript.PayToAddrScript(address)
			if err != nil {
				log.Printf("Failed to create scriptPubKey: %v", err)
				return nil, err
			}
			recipient.PkScript = scriptPubKey

			newRecipients = append(newRecipients, recipient)
			continue
		}

		spRecipients = append(spRecipients, &bip352.Recipient{
			SilentPaymentAddress: recipient.Address,
			Amount:               uint64(recipient.Amount),
		})
	}

	var mainnet bool
	if chainParams.Name == chaincfg.MainNetParams.Name {
		mainnet = true
	}

	if len(spRecipients) > 0 {
		err := bip352.SenderCreateOutputs(spRecipients, vins, mainnet, false)
		if err != nil {
			return nil, err
		}
	}

	for _, spRecipient := range spRecipients {
		newRecipients = append(newRecipients, ConvertSPRecipient(spRecipient))
	}

	// This case might not be realistic so the check could potentially be removed safely
	if len(recipients) != len(newRecipients) {
		// for some reason we have a different number of recipients after parsing them.
		return nil, fmt.Errorf("bad length of recipients got %d needed %d", len(newRecipients), len(recipients))
	}

	return newRecipients, nil
}

// sanityCheckRecipientsForSending
// checks whether any of the Recipients lacks the necessary information to construct the transaction.
// required for every recipient: Recipient.PkScript and Recipient.Amount
func sanityCheckRecipientsForSending(recipients []*coinselector.Recipient) error {
	for _, recipient := range recipients {
		if recipient.PkScript == nil || recipient.Amount == 0 {
			// if we choose a lot of logging in this module/program we could log the incomplete recipient here
			return fmt.Errorf("incomplete recipient %s", recipient.Address)
		}
	}
	return nil
}

func CreateUnsignedPsbt(recipients []*coinselector.Recipient, vins []*bip352.Vin) (*psbt.Packet, error) {
	var txOutputs []*wire.TxOut
	for _, recipient := range recipients {
		txOutputs = append(txOutputs, wire.NewTxOut(int64(recipient.Amount), recipient.PkScript))
	}

	var txInputs []*wire.TxIn
	for _, vin := range vins {
		hash, err := chainhash.NewHash(bip352.ReverseBytesCopy(vin.Txid[:]))
		if err != nil {
			return nil, err
		}
		prevOut := wire.NewOutPoint(hash, vin.Vout)
		txInputs = append(txInputs, wire.NewTxIn(prevOut, nil, nil))
	}

	unsignedTx := &wire.MsgTx{
		Version: 2,
		TxIn:    txInputs,
		TxOut:   txOutputs,
	}

	packet := &psbt.Packet{
		UnsignedTx: txsort.Sort(unsignedTx),
	}

	return packet, nil
}

// SignPsbt
// fails if inputs in packet have a different order than vins
func SignPsbt(packet *psbt.Packet, vins []*bip352.Vin) error {
	if len(packet.UnsignedTx.TxIn) != len(vins) {
		return fmt.Errorf("mismatch with txIns (%d) and vins (%d)", len(packet.UnsignedTx.TxIn), len(vins))
	}

	prevOutsForFetcher := make(map[wire.OutPoint]*wire.TxOut, len(vins))

	// simple map to find correct vin for prevOutsForFetcher
	vinMap := make(map[string]bip352.Vin, len(vins))
	for _, v := range vins {
		vinMap[fmt.Sprintf("%x:%d", v.Txid, v.Vout)] = *v
	}

	for i := 0; i < len(vins); i++ {
		outpoint := packet.UnsignedTx.TxIn[i].PreviousOutPoint
		key := fmt.Sprintf("%x:%d", bip352.ReverseBytesCopy(outpoint.Hash[:]), outpoint.Index)
		vin, ok := vinMap[key]
		if !ok {
			err := fmt.Errorf("a vin was not found in the map, should not happen. upstream error in psbt and vin selection and or construction")
			return err
		}
		prevOutsForFetcher[outpoint] = wire.NewTxOut(int64(vin.Amount), vin.ScriptPubKey)
	}

	multiFetcher := txscript.NewMultiPrevOutFetcher(prevOutsForFetcher)

	sigHashes := txscript.NewTxSigHashes(packet.UnsignedTx, multiFetcher)

	var pInputs []psbt.PInput

	for iOuter, input := range packet.UnsignedTx.TxIn {
		signatureHash, err := txscript.CalcTaprootSignatureHash(sigHashes, txscript.SigHashDefault, packet.UnsignedTx, iOuter, multiFetcher)
		if err != nil {
			panic(err)
		}

		pInput, err := matchAndSign(input, signatureHash, vins)
		if err != nil {
			panic(err)
		}

		pInputs = append(pInputs, pInput)

	}

	packet.Inputs = pInputs

	return nil
}

func matchAndSign(
	input *wire.TxIn,
	signatureHash []byte,
	vins []*bip352.Vin,
) (
	psbt.PInput,
	error,
) {
	var psbtInput psbt.PInput

	for _, vin := range vins {
		if bytes.Equal(input.PreviousOutPoint.Hash[:], bip352.ReverseBytesCopy(vin.Txid[:])) &&
			input.PreviousOutPoint.Index == vin.Vout {
			privKey, pk := btcec.PrivKeyFromBytes(vin.SecretKey[:])

			if pk.Y().Bit(0) == 1 {
				newBytes := privKey.Key.Negate().Bytes()
				privKey, _ = btcec.PrivKeyFromBytes(newBytes[:])
			}
			signature, err := schnorr.Sign(privKey, signatureHash)
			if err != nil {
				panic(err)
			}

			var witnessBytes bytes.Buffer
			err = psbt.WriteTxWitness(&witnessBytes, [][]byte{signature.Serialize()})
			if err != nil {
				panic(err)
			}

			return psbt.PInput{
				WitnessUtxo:        wire.NewTxOut(int64(vin.Amount), vin.ScriptPubKey),
				SighashType:        txscript.SigHashDefault,
				FinalScriptWitness: witnessBytes.Bytes(),
			}, err
		}
	}

	return psbtInput, fmt.Errorf("no matching vin found for txInput")

}
func ConvertSPRecipient(recipient *bip352.Recipient) *coinselector.Recipient {
	return &coinselector.Recipient{
		Address:  recipient.SilentPaymentAddress,
		PkScript: append([]byte{0x51, 0x20}, recipient.Output[:]...),
		Amount:   recipient.Amount,
	}
}

func ConvertOwnedUTXOIntoVin(utxo *scanwallet.OwnedUTXO) bip352.Vin {
	vin := bip352.Vin{
		Txid:         utxo.Txid,
		Vout:         utxo.Vout,
		Amount:       utxo.Amount,
		ScriptPubKey: append([]byte{0x51, 0x20}, utxo.PubKey[:]...),
		SecretKey:    &utxo.PrivKeyTweak,
		Taproot:      true,
	}
	return vin
}
