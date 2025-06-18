package engine

import (
	"context"
	"fmt"

	trx "github.com/bitcoin-sv/go-sdk/transaction"
	"github.com/bitcoin-sv/spv-wallet/engine/spverrors"
)

type internalIncomingTx struct {
	Tx *Transaction
}

func (strategy *internalIncomingTx) Name() string {
	return "internal_incoming_tx"
}

func (strategy *internalIncomingTx) Execute(ctx context.Context, c ClientInterface, _ []ModelOps) (*Transaction, error) {
	transaction := strategy.Tx

	logger := c.Logger()

	if _isTokenTransaction(transaction.parsedTx) {
		logger.Info().Str("strategy", "internal incoming").Msg("Token transaction FOUND")

		tm, err := buildTransferMessage(transaction)
		if err != nil {
			return nil, spverrors.ErrTokenValidationFailed.Wrap(err)
		}

		logger.Info().Str("strategy", "internal incoming").Any("transfer-data", tm).Msg("")

		err = c.Tokens().VerifyAndSaveTokenTransfer(ctx, transaction.Hex)
		if err != nil {
			return nil, spverrors.ErrTokenValidationFailed.Wrap(err)
		}
		logger.Info().Str("strategy", "internal incoming").Msg("Token transaction successfully VALIDATED")
	}

	if err := broadcastTransaction(ctx, transaction); err != nil {
		return nil, err
	}
	transaction.TxStatus = TxStatusBroadcasted

	if err := transaction.Save(ctx); err != nil {
		c.Logger().Error().Str("txID", transaction.ID).Err(err).Msg("Incoming internal transaction has been broadcasted but failed save to db")
	}

	return transaction, nil
}

func (strategy *internalIncomingTx) Validate() error {
	if strategy.Tx == nil {
		return spverrors.ErrEmptyTx
	}

	if _, err := trx.NewTransactionFromHex(strategy.Tx.Hex); err != nil {
		return spverrors.ErrInvalidHex
	}

	return nil // is valid
}

func (strategy *internalIncomingTx) TxID() string {
	return strategy.Tx.ID
}

func (strategy *internalIncomingTx) LockKey() string {
	return fmt.Sprintf("incoming-%s", strategy.Tx.ID)
}

// QUICK AND DIRTY
type TransferMessage struct {
	// envelope
	SenderID          string `json:"sender_id"`
	SenderChangeVouts []uint `json:"sender_vouts"`

	ReceiverID    string `json:"receiver_id"`
	ReceiverVouts []uint `json:"receiver_vouts"`

	FeeVouts []uint `json:"fee_vouts"`

	//CoinID string/*engine_bsv21.TokenID*/ `json:"coin_id"`

	Hex string `json:"hex"`
}

func buildTransferMessage(t *Transaction) (*TransferMessage, error) {
	draft, err := getDraftTransactionID(context.Background(), t.XPubID, t.DraftID, t.GetOptions(false)...)
	if err != nil {
		return nil, err
	}

	if draft == nil {
		return nil, spverrors.ErrCouldNotFindDraftTx
	}

	var senderOuts []uint
	var receiverOuts []uint
	var feeOuts []uint

	var transferOut *TransactionOutput

	for i, out := range draft.Configuration.Outputs {
		if !out.Token {
			continue
		}

		if out.TokenChange {
			senderOuts = append(senderOuts, uint(i))
			continue
		}

		if out.TokenFee {
			feeOuts = append(feeOuts, uint(i))
			continue
		}

		receiverOuts = append(receiverOuts, uint(i))
		transferOut = out // i know i assign it multiple times, it's ok for now
	}

	if transferOut == nil {
		return nil, spverrors.ErrInvalidTransferNoTransfer
	}

	return &TransferMessage{
		SenderID:          transferOut.PaymailP4.FromPaymail,
		ReceiverID:        fmt.Sprintf("%s@%s", transferOut.PaymailP4.Alias, transferOut.PaymailP4.Domain),
		SenderChangeVouts: senderOuts,
		ReceiverVouts:     receiverOuts,
		FeeVouts:          feeOuts,

		Hex: t.Hex,
	}, nil
}
