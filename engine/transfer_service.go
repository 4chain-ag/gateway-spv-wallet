package engine

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bitcoin-sv/go-paymail"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

// Intent is the model for validating a transfer request
type Intent struct {
	SenderID     string     `json:"senderId" example:"alice@spv-wallet.com"`
	ReceiverID   string     `json:"receiverId" example:"bob@spv-wallet.com"`
	Nonce        string     `json:"nonce" example:"1234567890abcdef"`
	StablecoinID string     `json:"stablecoinId" example:"0761072ea3519adcbf4c2b9061bf64cb52243533f72d1cec47280a6eabfb3ad5_0"`
	Amount       uint64     `json:"amount" example:"1000000"`
	Banknotes    []Banknote `json:"banknotes"`
	Metadata     Metadata   `json:"metadata" swaggertype:"object,string" example:"key:value,key2:value2"`
}

// ValidationResponse is the model for the response of a transfer intent validation
type ValidationResponse struct {
	Nonce   string               `json:"nonce" example:"1234567890abcdef"`
	Outputs []*TransactionOutput `json:"outputs"`
}

// TransferService provides methods to validate and send transfer intents
type TransferService struct {
	log       *zerolog.Logger
	validator IntentValidator
}

// NewTransferService creates a new instance of TransferService with the provided validator and logger
func NewTransferService(validator IntentValidator, log *zerolog.Logger) *TransferService {
	return &TransferService{
		log:       log,
		validator: validator,
	}
}

// ValidateIntent validates transfer intent by checking the sender, calculating transaction outputs, and creating a transfer intent model
func (s *TransferService) ValidateIntent(ctx context.Context, c ClientInterface, intent *Intent) (*ValidationResponse, error) {
	s.log.Debug().Str("senderID", intent.SenderID).Msg("Validating transfer intent")

	if err := s.validator.ValidateSender(ctx, intent.SenderID); err != nil {
		s.log.Error().Err(err).Str("senderID", intent.SenderID).Msg("Sender validation failed")
		return nil, err
	}

	txOutputs, feeOutputs, err := s.validator.GetTxOutputs(ctx, intent)
	if err != nil {
		s.log.Error().Err(err).Str("senderID", intent.SenderID).Msg("Failed to get fee outputs")
		return nil, err
	}

	opts := []ModelOps{WithClient(c)}
	outputs := append(txOutputs, feeOutputs...)
	transferIntent, err := CreateTransferIntent(intent, outputs, c.DefaultModelOptions(append(opts, New())...)...)
	if err != nil {
		s.log.Error().Err(err).Str("senderID", intent.SenderID).Msg("Failed to create transfer intent")
		return nil, fmt.Errorf("failed to create transfer intent: %w", err)
	}

	err = transferIntent.Save(ctx)
	if err != nil {
		s.log.Error().Err(err).Str("senderID", intent.SenderID).Msg("Failed to save transfer intent")
		return nil, fmt.Errorf("failed to save transfer intent: %w", err)
	}

	return &ValidationResponse{
		Nonce:   transferIntent.Nonce,
		Outputs: outputs,
	}, nil
}

// SendTransferIntent sends the transfer intent to the receiver's paymail server for validation
func (s *TransferService) SendTransferIntent(intent Intent) (*ValidationResponse, error) {
	_, domain, _ := paymail.SanitizePaymail(intent.ReceiverID)
	path := "bsvalias/transfer-intent"
	url := fmt.Sprintf("https://%s/%s", domain, path)

	httpClient := resty.New()
	resp, err := httpClient.R().
		SetBody(intent).
		Post(url)

	if err != nil {
		s.log.Error().Err(err).Str("receiverID", intent.ReceiverID).Msg("Failed to send transfer intent")
		return nil, fmt.Errorf("failed to send transfer intent: %w", err)
	}

	var vr ValidationResponse
	if err = json.Unmarshal(resp.Body(), &vr); err != nil {
		s.log.Error().Err(err).Str("receiverID", intent.ReceiverID).Msg("Failed to unmarshal validation response")
		return nil, fmt.Errorf("failed to unmarshal validation response: %w", err)
	}

	return &vr, nil
}
