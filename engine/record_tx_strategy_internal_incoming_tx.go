package engine

import (
	"context"
	"fmt"

	api "github.com/4chain-AG/gateway-overlay/pkg/open_api"
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

		err = c.Tokens().VerifyAndSaveTokenTransfer(ctx, tm)
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

func buildTransferMessage(t *Transaction) (*api.PutApiV1Bsv21TransferJSONRequestBody, error) {
	draft, err := getDraftTransactionID(context.Background(), t.XPubID, t.DraftID, t.GetOptions(false)...)
	if err != nil {
		return nil, err
	}

	if draft == nil {
		return nil, spverrors.ErrCouldNotFindDraftTx
	}

	var senderOuts []int
	var receiverOuts []int
	var feeOuts []int

	var transferOut *TransactionOutput

	for i, out := range draft.Configuration.Outputs {
		if !out.Token {
			continue
		}

		if out.TokenChange {
			senderOuts = append(senderOuts, i)
			continue
		}

		if out.TokenFee {
			feeOuts = append(feeOuts, i)
			continue
		}

		receiverOuts = append(receiverOuts, i)
		transferOut = out // i know i assign it multiple times, it's ok for now
	}

	if transferOut == nil {
		return nil, spverrors.ErrInvalidTransferNoTransfer
	}

	return &api.PutApiV1Bsv21TransferJSONRequestBody{
		SenderId:   transferOut.PaymailP4.FromPaymail,
		ReceiverId: fmt.Sprintf("%s@%s", transferOut.PaymailP4.Alias, transferOut.PaymailP4.Domain),

		SenderVouts:   &senderOuts,
		ReceiverVouts: &receiverOuts,
		FeeVouts:      &feeOuts,

		Hex: t.Hex,
	}, nil
}

func ptrTo[T any](v T) *T {
	return &v
}
