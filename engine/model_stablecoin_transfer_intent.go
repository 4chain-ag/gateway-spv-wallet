package engine

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	crypto "github.com/bitcoin-sv/go-sdk/primitives/hash"
	"github.com/bitcoin-sv/spv-wallet/engine/datastore"
	"github.com/bitcoin-sv/spv-wallet/engine/spverrors"
	"gorm.io/gorm"
)

// StablecoinTransferIntent is the model for validating a transfer request
type StablecoinTransferIntent struct {
	// Base model
	Model

	ID           string             `json:"id" toml:"id" yaml:"id" gorm:"primaryKey;type:char(64);not null;uniqueIndex" example:"0761072ea3519adcbf4c2b9061bf64cb52243533f72d1cec47280a6eabfb3ad5"`
	SenderID     string             `json:"senderId" example:"example@spv-wallet.com" toml:"senderId" yaml:"senderId" gorm:"<-;comment:Sender identifier"`
	ReceiverID   string             `json:"receiverId" example:"example@spv-wallet.com" toml:"receiverId" yaml:"receiverId" gorm:"<-;comment:Receiver identifier"`
	Nonce        string             `json:"nonce" example:"1234567890abcdef" toml:"nonce" yaml:"nonce" gorm:"<-;comment:Nonce for the transfer request"`
	StablecoinID string             `json:"stablecoinId" example:"0761072ea3519adcbf4c2b9061bf64cb52243533f72d1cec47280a6eabfb3ad5_0" toml:"stablecoinId" yaml:"stablecoinId" gorm:"<-;comment:Stablecoin identifier"`
	Amount       uint64             `json:"amount" example:"1000000" toml:"amount" yaml:"amount" gorm:"<-;comment:Amount of tokens to be transferred"`
	Banknotes    Banknotes          `json:"banknotes" toml:"banknotes" yaml:"banknotes" gorm:"<-;type:json;comment:List of banknotes involved in the transfer"`
	Outputs      TransactionOutputs `json:"outputs" toml:"outputs" yaml:"outputs" gorm:"<-;type:json;comment:List of outputs involved in the transfer"`
}

// BeforeCreating is a hook that is called before the model is created in the database
func (m *StablecoinTransferIntent) BeforeCreating(_ context.Context) (err error) {
	m.Client().Logger().Debug().
		Str("draftTxID", m.GetID()).
		Msgf("starting: %s BeforeCreating hook...", m.Name())

	m.Client().Logger().Debug().
		Str("draftTxID", m.GetID()).
		Msgf("end: %s BeforeCreating hook", m.Name())
	return
}

// GetModelName returns the name of the model
func (m *StablecoinTransferIntent) GetModelName() string {
	return ModelTransferIntent.String()
}

// GetModelTableName returns the table name for the model
func (m *StablecoinTransferIntent) GetModelTableName() string {
	return tableStablecoinTransferIntents
}

// PostMigrate is called after the model is migrated to the database
func (m *StablecoinTransferIntent) PostMigrate(client datastore.ClientInterface) error {
	err := client.IndexMetadata(client.GetTableName(tableStablecoinTransferIntents), metadataField)
	return spverrors.Wrapf(err, "failed to index metadata column on model %s", m.GetModelName())
}

// Save will save the model into the Datastore
func (m *StablecoinTransferIntent) Save(ctx context.Context) error {
	return Save(ctx, m)
}

// CreateStablecoinTransferIntent creates a new StablecoinTransferIntent with the provided parameters
func CreateStablecoinTransferIntent(intent *Intent, outputs []*TransactionOutput, opts ...ModelOps) (*StablecoinTransferIntent, error) {
	if intent == nil {
		return nil, errors.New("transfer intent cannot be nil")
	}

	nonce, err := generateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	computedNonces := fmt.Sprintf("%s%s", intent.Nonce, nonce)
	hash := crypto.Sha256([]byte(computedNonces))
	refID := hex.EncodeToString(hash[:])

	sti := &StablecoinTransferIntent{
		ID:           refID,
		SenderID:     intent.SenderID,
		ReceiverID:   intent.ReceiverID,
		Nonce:        nonce,
		StablecoinID: intent.StablecoinID,
		Amount:       intent.Amount,
		Banknotes:    intent.Banknotes,
		Outputs:      outputs,

		Model: *NewBaseModel(ModelTransferIntent, opts...),
	}

	return sti, nil
}

func generateNonce() (string, error) {
	bb := make([]byte, 32)
	_, err := rand.Read(bb)
	if err != nil {
		return "", fmt.Errorf("failed to read bytes after rand: %w", err)
	}

	return hex.EncodeToString(bb), nil
}

func getStablecoinTransferIntentByID(ctx context.Context, id string, opts ...ModelOps) (*StablecoinTransferIntent, error) {
	conditions := map[string]interface{}{
		idField: id,
	}

	sti := &StablecoinTransferIntent{Model: *NewBaseModel(
		ModelTransferIntent,
		opts...,
	)}
	if err := Get(ctx, sti, conditions, false, defaultDatabaseReadTimeout, true); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return sti, nil
}
