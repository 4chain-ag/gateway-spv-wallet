package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/4chain-AG/gateway-overlay/pkg/token_engine/bsv21"
	"github.com/bitcoin-sv/go-paymail"
	"github.com/bitcoin-sv/go-sdk/script"
	trx "github.com/bitcoin-sv/go-sdk/transaction"
	"github.com/bitcoin-sv/spv-wallet/engine/spverrors"
	"github.com/bitcoin-sv/spv-wallet/engine/utils"
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

type Transfer struct {
	RefID string `json:"refId" example:"0761072ea3519adcbf4c2b9061bf64cb52243533f72d1cec47280a6eabfb3ad5"`
	TxHex string `json:"txHex" example:"0100000001..."`

	// SpecialOperation is used to indicate special operations like issue or redeem
	SpecialOperation string `json:"_,omitempty" example:"issue/redeem"`
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

// ValidateTransfer validates the transfer by comparing the scripts in the transfer intent with the transaction outputs
func (s *TransferService) ValidateTransfer(c ClientInterface, transfer Transfer) error {
	tx, err := trx.NewTransactionFromHex(transfer.TxHex)
	if err != nil {
		return spverrors.ErrInvalidHex
	}

	transferIntent, err := getTransferIntentByID(context.Background(), transfer.RefID, c.DefaultModelOptions()...)
	if err != nil {
		return fmt.Errorf("error getting transfer intent: %w", err)
	}

	err = compareScripts(transferIntent, tx, transfer)
	if err != nil {
		s.log.Error().Err(err).Str("refID", transfer.RefID).Msg("Transfer validation failed")
		return fmt.Errorf("transfer validation failed: %w", err)
	}

	return nil
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

// SendTransfer sends the transfer to the receiver's paymail server for validation
func (s *TransferService) SendTransfer(receiverDomain string, transfer Transfer) error {
	path := "bsvalias/transfer"
	url := fmt.Sprintf("https://%s/%s", receiverDomain, path)

	httpClient := resty.New()
	resp, err := httpClient.R().
		SetBody(transfer).
		Post(url)

	if err != nil {
		s.log.Error().Err(err).Str("receiverDomain", receiverDomain).Msg("Failed to send transfer")
		return fmt.Errorf("failed to send transfer: %w", err)
	}

	if resp.StatusCode() != 200 {
		s.log.Error().Str("receiverDomain", receiverDomain).Int("statusCode", resp.StatusCode()).
			Msg("Failed to send transfer, received non-200 status code")
		return fmt.Errorf("failed to send transfer, received status code: %d", resp.StatusCode())
	}

	return nil
}

// IncomingTransfer processes an incoming transfer by validating it, creating a transaction from the hex, and recording it
func (s *TransferService) IncomingTransfer(ctx context.Context, c ClientInterface, transfer Transfer) (*Transaction, error) {
	err := s.ValidateTransfer(c, transfer)
	if err != nil {
		s.log.Error().Err(err).Str("refID", transfer.RefID).Msg("Transfer validation failed")
		return nil, spverrors.Wrapf(err, "transfer validation failed")
	}

	sdkTx, err := trx.NewTransactionFromHex(transfer.TxHex)
	if err != nil {
		return nil, spverrors.Wrapf(err, "transfer validation failed, cannot create transaction from hex")
	}

	rts, err := getIncomingTxRecordStrategy(ctx, c, sdkTx, utils.IsEf(transfer.TxHex))
	if err != nil {
		s.log.Error().Err(err).Str("refID", transfer.RefID).Msg("Failed to get incoming transaction record strategy")
		return nil, spverrors.Wrapf(err, "failed to get incoming transaction record strategy")
	}

	transaction, err := recordTransaction(ctx, c, rts, WithMetadatas(map[string]interface{}{"transfer": transfer}))
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

// NotifyGatewayAboutTransfer is a placeholder method for notifying the gateway about the transfer
func (s *TransferService) NotifyGatewayAboutTransfer() {
	// This method is a placeholder for notifying the gateway about the transfer.
	// The implementation will depend on the specific requirements and architecture of the system.
	s.log.Info().Msg("NotifyGatewayAboutTransfer method called, but not implemented yet.")
}

// compareScripts compares the scripts in the transfer intent with the transaction outputs
// it is required for tx to contain all outputs from the intent
func compareScripts(intent *TransferIntent, tx *trx.Transaction, transfer Transfer) error {
	// intent can be nil if the transaction is issue or redeem operation
	if intent == nil {
		if transfer.SpecialOperation != "" {
			return nil
		}
		return fmt.Errorf("transfer intent is nil, cannot compare scripts")
	}

	for _, out := range intent.Outputs {
		intentScript, err := script.NewFromHex(out.Script)
		if err != nil {
			return fmt.Errorf("failed to create script from hex: %w", err)
		}

		intentOperation, err := getTokenOperationFromScript(intentScript)
		if err != nil {
			return fmt.Errorf("failed to get token operation from script: %w", err)
		}

		found := false
		for _, txOut := range tx.Outputs {
			txOperation, err := getTokenOperationFromScript(txOut.LockingScript)
			if err != nil {
				continue
			}

			if !reflect.DeepEqual(intentOperation, txOperation) {
				continue
			}

			found = true
		}
		if !found {
			return fmt.Errorf("output %s not found in transaction outputs", out.Script)
		}
	}

	return nil
}

func getTokenOperationFromScript(s *script.Script) (*bsv21.TokenOperation, error) {
	inscription, err := bsv21.FindInscription(s)
	if err != nil || inscription == nil {
		return nil, fmt.Errorf("failed to find inscription in script: %w", err)
	}

	inscriptionData, err := bsv21.NewFromInscription("id", 1, inscription)
	if err != nil {
		return nil, fmt.Errorf("failed to create inscription data from inscription: %w", err)
	}

	return inscriptionData, nil
}
