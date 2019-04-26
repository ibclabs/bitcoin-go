package bitcoin

import (
	"bytes"
	"encoding/hex"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

type Transaction struct {
	TxId          string `json:"txid"`
	SourceAddress string `json:"source_address"`
	UnsignedTx    string `json:"unsignedtx"`
	SignedTx      string `json:"signedtx"`
}

func GetPublicKey(wif *btcutil.WIF, compress bool) []byte {
	if compress {
		return wif.PrivKey.PubKey().SerializeCompressed()
	}
	return wif.PrivKey.PubKey().SerializeUncompressed()
}

type Destination struct {
	Addr   string
	Amount int64
}

func CreateTransactionNew(secret string, dests []Destination, inputTxHashes []string, compress bool) (Transaction, error) {
	var transaction Transaction

	wif, err := btcutil.DecodeWIF(secret)
	if err != nil {
		return transaction, err
	}

	serialized := GetPublicKey(wif, compress)

	addressPubKey, err := btcutil.NewAddressPubKey(serialized, &chaincfg.MainNetParams)
	if err != nil {
		return transaction, err
	}

	sourceTx := wire.NewMsgTx(wire.TxVersion)
	for i, inputTxHash := range inputTxHashes {
		sourceUTXOHash, err := chainhash.NewHashFromStr(inputTxHash)
		if err != nil {
			return transaction, err
		}
		sourceUTXO := wire.NewOutPoint(sourceUTXOHash, uint32(i))
		sourceTxIn := wire.NewTxIn(sourceUTXO, nil, nil)

		sourceTx.AddTxIn(sourceTxIn)
	}

	sourceAddress, err := btcutil.DecodeAddress(addressPubKey.EncodeAddress(), &chaincfg.MainNetParams)
	if err != nil {
		return transaction, err
	}

	transaction.SourceAddress = sourceAddress.EncodeAddress()

	sourcePkScript, err := txscript.PayToAddrScript(sourceAddress)
	if err != nil {
		return transaction, err
	}

	outpoints := make([]*wire.TxOut, len(dests))

	for i, destination := range dests {
		sourceTxOut := wire.NewTxOut(destination.Amount, sourcePkScript)
		sourceTx.AddTxOut(sourceTxOut)

		outpoints[i] = sourceTxOut
	}

	sourceTxHash := sourceTx.TxHash()
	transaction.TxId = sourceTxHash.String()

	redeemTx := wire.NewMsgTx(wire.TxVersion)
	for i, destination := range dests {
		addr, err := btcutil.DecodeAddress(destination.Addr, &chaincfg.MainNetParams)
		if err != nil {
			return transaction, err
		}
		destPkScript, err := txscript.PayToAddrScript(addr)
		if err != nil {
			return transaction, err
		}

		prevOut := wire.NewOutPoint(&sourceTxHash, uint32(i))
		redeemTxIn := wire.NewTxIn(prevOut, nil, nil)
		redeemTx.AddTxIn(redeemTxIn)
		redeemTxOut := wire.NewTxOut(destination.Amount, destPkScript)
		redeemTx.AddTxOut(redeemTxOut)
	}

	for i, out := range outpoints {
		sigScript, err := txscript.SignatureScript(redeemTx, i, out.PkScript, txscript.SigHashSingle, wif.PrivKey, compress)
		if err != nil {
			return transaction, err
		}

		redeemTx.TxIn[i].SignatureScript = sigScript
	}

	var unsignedTx bytes.Buffer
	err = sourceTx.Serialize(&unsignedTx)
	if err != nil {
		return transaction, err
	}

	transaction.UnsignedTx = hex.EncodeToString(unsignedTx.Bytes())

	var signedTx bytes.Buffer
	err = redeemTx.Serialize(&signedTx)
	if err != nil {
		return transaction, err
	}

	transaction.SignedTx = hex.EncodeToString(signedTx.Bytes())

	return transaction, nil
}
