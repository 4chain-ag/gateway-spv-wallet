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
	feeAmount := uint64(0)
	feeIssuer := ""

	// Check if the intent has a metadata key that indicates fee-free transactions
	if _, ok := intent.Metadata[TransactionFeeFreeKey]; !ok {
		rules, err := d.c.GatewayClient().GetStablecoinRules(intent.StablecoinID)
		if err != nil {
			return nil, nil, err
		}

		if intent.ReceiverID != rules.EmitterID {
			feeIssuer, feeAmount = d.getApplicableFee(rules.Fees, intent.Amount)
			if intent.Amount <= feeAmount {
				return nil, nil, errors.New("fee will cover all of the transfer")
			}
		}
	}

	// Sort banknotes to consume larger ones first for efficiency
	sort.Slice(intent.Banknotes, func(i, j int) bool {
		return intent.Banknotes[i].Amount > intent.Banknotes[j].Amount
	})

	// Create a mutable copy of banknotes
	remainingBanknotes := make([]Banknote, len(intent.Banknotes))
	copy(remainingBanknotes, intent.Banknotes)

	// First, try to create fee outputs from the main stablecoin banknotes
	// This function will also return the remaining banknotes
	feeOutputs, remainingBanknotes, err = d.createFeeOutputsFromBanknotes(remainingBanknotes, feeIssuer, feeAmount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create fee outputs: %w", err)
	}

	// Now, create the outputs for the receiver from the remaining banknotes
	txOutputs, err = d.createReceiverOutputs(remainingBanknotes, intent.ReceiverID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create receiver outputs: %w", err)
	}

	return
}

// createFeeOutputsFromBanknotes consumes banknotes to create outputs for a specific recipient
// and returns the outputs and any remaining banknotes.
func (d defaultValidator) createFeeOutputsFromBanknotes(banknotes []Banknote, to string, requiredAmount uint64) ([]*TransactionOutput, []Banknote, error) {
	outputs := make([]*TransactionOutput, 0)
	consumedAmount := uint64(0)
	remainingBanknotes := make([]Banknote, 0)

	for _, b := range banknotes {
		if consumedAmount >= requiredAmount {
			remainingBanknotes = append(remainingBanknotes, b)
			continue
		}

		// Calculate how much of the current banknote to use
		amountToUse := b.Amount
		if consumedAmount+amountToUse > requiredAmount {
			amountToUse = requiredAmount - consumedAmount
		}

		// Create output fee for the amount to be used
		output, err := d.createTransactionOutput(b.Serial, to, amountToUse, true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create output for %s: %w", b.Serial, err)
		}
		outputs = append(outputs, output)
		consumedAmount += amountToUse

		// Check if there's any remaining value in the current banknote
		if b.Amount > amountToUse {
			remainingBanknotes = append(remainingBanknotes, Banknote{Amount: b.Amount - amountToUse, Serial: b.Serial})
		}
	}

	if consumedAmount < requiredAmount {
		return nil, nil, fmt.Errorf("not enough banknote value to cover the required amount")
	}

	return outputs, remainingBanknotes, nil
}

// createReceiverOutputs creates outputs for the receiver from the remaining banknotes.
func (d defaultValidator) createReceiverOutputs(banknotes []Banknote, receiverID string) ([]*TransactionOutput, error) {
	outputs := make([]*TransactionOutput, 0, len(banknotes))
	for _, b := range banknotes {
		output, err := d.createTransactionOutput(b.Serial, receiverID, b.Amount, false)
		if err != nil {
			return nil, fmt.Errorf("failed to create receiver output: %w", err)
		}
		outputs = append(outputs, output)
	}
	return outputs, nil
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

// createTransactionOutput is a helper function to create a single TransactionOutput.
func (d defaultValidator) createTransactionOutput(serial, to string, amount uint64, isTokenFee bool) (*TransactionOutput, error) {
	txScript, err := bsv21.NewBsv21Transfer(bsv21.TokenID(serial), amount)
	if err != nil {
		d.log.Error().Err(err).Msg("Failed to create BSV21 transfer script")
		return nil, fmt.Errorf("failed to create BSV21 transfer script: %w", err)
	}

	return &TransactionOutput{
		Satoshis: 1,
		Script:   txScript.String(),
		To:       to,
		Token:    true,
		TokenFee: isTokenFee,
	}, nil
}
