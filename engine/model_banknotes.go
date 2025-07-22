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

// Banknote represents a banknote in the transfer request
type Banknote struct {
	Amount uint64 `json:"value" example:"1000000" toml:"value" yaml:"value" gorm:"<-;comment:Value of the banknote"`
	Serial string `json:"serial" example:"0761072ea3519adcbf4c2b9061bf64cb52243533f72d1cec47280a6eabfb3ad5_0" toml:"serial" yaml:"serial" gorm:"<-;comment:Serial number of the banknote"`
}

// Banknotes is a slice of Banknote, used for JSON serialization
type Banknotes []Banknote

// GormDataType type in gorm
func (i Banknotes) GormDataType() string {
	return gormTypeText
}

// Scan scan value into JSON, implements sql.Scanner interface
func (i *Banknotes) Scan(value interface{}) error {
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
func (i Banknotes) Value() (driver.Value, error) {
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
func (Banknotes) GormDBDataType(db *gorm.DB, _ *schema.Field) string {
	if db.Dialector.Name() == datastore.Postgres {
		return datastore.JSONB
	}
	return datastore.JSON
}
