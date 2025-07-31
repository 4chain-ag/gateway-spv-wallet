package engine

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/bitcoin-sv/spv-wallet/engine/datastore"
	"github.com/bitcoin-sv/spv-wallet/engine/spverrors"
	"github.com/bitcoin-sv/spv-wallet/engine/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// TransactionOutput is an output on the transaction config
type TransactionOutput struct {
	OpReturn     *OpReturn       `json:"op_return,omitempty" toml:"op_return" yaml:"op_return"`                // Add op_return data as an output
	PaymailP4    *PaymailP4      `json:"paymail_p4,omitempty" toml:"paymail_p4" yaml:"paymail_p4"`             // Additional information for P4 or Paymail
	Satoshis     uint64          `json:"satoshis" toml:"satoshis" yaml:"satoshis"`                             // Set the specific satoshis to send (when applicable)
	Script       string          `json:"script,omitempty" toml:"script" yaml:"script"`                         // custom (non-standard) script output
	Scripts      []*ScriptOutput `json:"scripts" toml:"scripts" yaml:"scripts"`                                // Add script outputs
	To           string          `json:"to,omitempty" toml:"to" yaml:"to"`                                     // To address, paymail, handle
	UseForChange bool            `json:"use_for_change,omitempty" toml:"use_for_change" yaml:"use_for_change"` // if set, no change destinations will be created, but all outputs flagged will get the change

	Token       bool `json:"token"`
	TokenChange bool `json:"token_change"`
	TokenFee    bool `json:"token_fee"`
}

// TransactionOutputs is a slice of TransactionOutput, used for JSON serialization
type TransactionOutputs []*TransactionOutput

// GormDataType type in gorm
func (i TransactionOutputs) GormDataType() string {
	return gormTypeText
}

// Scan scan value into JSON, implements sql.Scanner interface
func (i *TransactionOutputs) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	byteValue, err := utils.ToByteArray(value)
	if err != nil {
		return nil
	}

	err = json.Unmarshal(byteValue, &i)
	return spverrors.Wrapf(err, "failed to parse IDs from JSON, data: %v", value)
}

// Value return json value, implement driver.Valuer interface
func (i TransactionOutputs) Value() (driver.Value, error) {
	if i == nil {
		return nil, nil
	}
	marshal, err := json.Marshal(i)
	if err != nil {
		return nil, spverrors.Wrapf(err, "failed to convert IDs to JSON, data: %v", i)
	}

	return string(marshal), nil
}

// GormDBDataType the gorm data type for metadata
func (TransactionOutputs) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	if db.Dialector.Name() == datastore.Postgres {
		return datastore.JSONB
	}
	return datastore.JSON
}
