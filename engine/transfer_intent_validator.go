package engine

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/4chain-AG/gateway-overlay/pkg/token_engine/bsv21"
	"github.com/bitcoin-sv/spv-wallet/engine/gateway"
	"github.com/rs/zerolog"
)

// IntentValidator defines the interface for validating intents and retrieving transaction outputs.
type IntentValidator interface {
	ValidateSender(ctx context.Context, senderID string) error
	GetTxOutputs(_ context.Context, intent *Intent) (txOutputs []*TransactionOutput, feeOutputs []*TransactionOutput, err error)
}

type defaultValidator struct {
	log *zerolog.Logger
	c   ClientInterface
}

// NewDefaultIntentValidator creates a new instance of the default intent validator.
func NewDefaultIntentValidator(log *zerolog.Logger, c ClientInterface) IntentValidator {
	l := log.With().Str("component", "default-intent-validator").Logger()
	return &defaultValidator{
		log: &l,
		c:   c,
	}
}

// ValidateSender validates the sender ID. Currently, it does not implement any validation logic and always returns nil.
func (d defaultValidator) ValidateSender(_ context.Context, senderID string) error {
	d.log.Debug().Str("senderID", senderID).Msg("Validating sender ID")
	return nil // No validation logic implemented, always returns nil
}

// GetTxOutputs Creates transaction outputs for the intent, including fee outputs if applicable.
func (d defaultValidator) GetTxOutputs(_ context.Context, intent *Intent) (txOutputs []*TransactionOutput, feeOutputs []*TransactionOutput, err error) {
	d.log.Debug().Str("senderID", intent.SenderID).Msg("Getting fee outputs for intent")
	txOutputs, feeOutputs, err = d.handleStablecoinOutputs(intent)
	if err != nil {
		d.log.Error().Err(err).Str("senderID", intent.SenderID).Msg("Failed to handle stablecoin fee")
		return nil, nil, fmt.Errorf("failed to handle stablecoin fee: %w", err)
	}
	return
}

func (d defaultValidator) handleStablecoinOutputs(intent *Intent) (txOutputs []*TransactionOutput, feeOutputs []*TransactionOutput, err error) {
	rules, err := d.c.GatewayClient().GetStablecoinRules(intent.StablecoinID)
	if err != nil {
		return nil, nil, err
	}

	feeAmount := uint64(0)
	// check if the intent has a metadata key that indicates fee-free transactions
	if _, ok := intent.Metadata[TransactionFeeFreeKey]; !ok {
		feeAmount, feeOutputs, err = d.calculateAndCreateFeeOutputs(intent, rules)
		if err != nil {
			d.log.Error().Err(err).Str("senderID", intent.SenderID).Msg("Failed to calculate fee outputs")
			return nil, nil, fmt.Errorf("failed to calculate fee outputs: %w", err)
		}
	}

	// TODO: fee will have specific serial number, so we need to decrease the amount of banknotes
	txOutputs, err = d.createNewOutputs(intent, feeAmount)
	if err != nil {
		d.log.Error().Err(err).Str("senderID", intent.SenderID).Msg("Failed to create new outputs")
		return nil, nil, fmt.Errorf("failed to create new outputs: %w", err)
	}

	return
}

func (d defaultValidator) calculateAndCreateFeeOutputs(intent *Intent, rules *gateway.StablecoinRule) (uint64, []*TransactionOutput, error) {
	feeOutputs := make([]*TransactionOutput, 0)
	feeAmount := uint64(0)
	if intent.ReceiverID != rules.EmitterID {
		var feeIssuer string
		feeIssuer, feeAmount = d.getApplicableFee(rules.Fees, intent.Amount)
		if feeAmount > 0 {
			if intent.Amount <= feeAmount {
				return 0, nil, errors.New("fee will cover all of the transfer")
			}

			var err error
			feeOutputs, err = d.createFeeOutputs(intent.StablecoinID, feeIssuer, feeAmount)
			if err != nil {
				return 0, nil, fmt.Errorf("failed to create fee output: %w", err)
			}
		}
	}
	return feeAmount, feeOutputs, nil
}

func (d defaultValidator) getApplicableFee(fees []*gateway.StablecoinFee, transactionAmount uint64) (string, uint64) {
	for _, fee := range fees {
		if transactionAmount >= fee.From && transactionAmount < fee.To {
			if fee.Type == "fixed" {
				return fee.CommissionRecipient, uint64(fee.Value)
			}

			if fee.Type == "percentage" {
				v := float64(transactionAmount) * (fee.Value / 100.0)
				return fee.CommissionRecipient, uint64(math.Ceil(v)) // TODO: think if it should be Math.Floor
			}

			panic(fmt.Sprintf("unknown fee type: %s", fee.Type))
		}
	}

	return "", 0
}

func (d defaultValidator) createFeeOutputs(stablecoinID string, feeIssuer string, feeAmount uint64) ([]*TransactionOutput, error) {
	sendFeeScript, err := bsv21.NewBsv21Transfer(bsv21.TokenID(stablecoinID), feeAmount)
	if err != nil {
		return nil, err
	}

	feeOutput := &TransactionOutput{
		Satoshis: 1,
		Script:   sendFeeScript.String(),
		To:       feeIssuer,

		Token:    true,
		TokenFee: true,
	}

	return []*TransactionOutput{feeOutput}, nil
}

func (d defaultValidator) createNewOutputs(intent *Intent, feeAmount uint64) ([]*TransactionOutput, error) {
	banknotes := intent.Banknotes
	sort.Slice(banknotes, func(i, j int) bool {
		return banknotes[i].Amount > banknotes[j].Amount
	})

	outputs := make([]*TransactionOutput, 0, len(banknotes))
	for _, b := range banknotes {
		// TODO: fee will have serial number and only the same one should be decreased
		// fee can consume more than one banknote, so we need to check if the feeAmount is greater than 0
		if b.Amount <= feeAmount {
			feeAmount -= b.Amount
			continue
		}

		remainingAmount := b.Amount - feeAmount
		txScript, err := bsv21.NewBsv21Transfer(bsv21.TokenID(intent.StablecoinID), remainingAmount)
		if err != nil {
			d.log.Error().Err(err).Msg("Failed to create BSV21 transfer script")
			return nil, fmt.Errorf("failed to create BSV21 transfer script: %w", err)
		}

		output := &TransactionOutput{
			Satoshis: 1, // minimal satoshi amount
			Script:   txScript.String(),
			To:       intent.ReceiverID,

			Token:    true,
			TokenFee: false,
		}

		outputs = append(outputs, output)
	}

	return outputs, nil
}
