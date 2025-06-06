package sql

import (
	"database/sql"
	"fmt"

	"github.com/bitcoin-sv/spv-wallet/engine/v2/database"
	"github.com/bitcoin-sv/spv-wallet/models/bsv"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type selectedUTXO struct {
	TxID               string
	Vout               uint32
	CustomInstructions datatypes.JSONSlice[bsv.CustomInstruction] `gorm:"column:custom_instructions"`
	Change             uint64
}

type inputsQueryComposer struct {
	userID              string
	outputsTotalValue   bsv.Satoshis
	txWithoutInputsSize uint64
	feeUnit             bsv.FeeUnit
}

func (c *inputsQueryComposer) build(db *gorm.DB) *gorm.DB {
	utxoTab := c.utxos(db)
	utxoWithChange := c.addChangeValueCalculation(db, utxoTab)
	utxoWithMinChange := c.searchForMinimalChangeValue(db, utxoWithChange)
	selectedOutpoints := c.chooseInputsToCoverOutputsAndFeesAndHaveMinimalChange(db, utxoWithMinChange)

	res := db.Model(&database.UserUTXO{}).Select(
		"ux."+txIdColumn,
		"ux."+voutColumn,
		"ux."+customInstructionsColumn,
		"sel."+minChange+" as change",
	).InnerJoins("ux join (?) sel ON sel.tx_id = ux.tx_id AND sel.vout = ux.vout", selectedOutpoints)
	return res
}

func (c *inputsQueryComposer) utxos(db *gorm.DB) *gorm.DB {
	return db.Model(&database.UserUTXO{}).
		Select(
			txIdColumn,
			voutColumn,
			c.remainingValue(),
			c.feeCalculatedWithoutChangeOutput(),
			c.feeCalculatedWithChangeOutput(),
		).
		Where("user_id = @userId", sql.Named("userId", c.userID))
}

func (c *inputsQueryComposer) addChangeValueCalculation(db *gorm.DB, utxoTab *gorm.DB) *gorm.DB {
	return db.Select(txIdColumn, voutColumn,
		"case when remaining_value - fee_no_change_output <= 0 then remaining_value - fee_no_change_output else remaining_value - fee_with_change_output end as change",
	).
		Table("(?) as utxo", utxoTab)
}

func (c *inputsQueryComposer) chooseInputsToCoverOutputsAndFeesAndHaveMinimalChange(db *gorm.DB, utxoWithMinChange *gorm.DB) *gorm.DB {
	return db.Select(txIdColumn, voutColumn, minChange).
		Table("(?) as utxoWithMinChange", utxoWithMinChange).
		Where("change <= " + minChange).
		Where("min_change is not null")
}

func (c *inputsQueryComposer) searchForMinimalChangeValue(db *gorm.DB, utxoWithChange *gorm.DB) *gorm.DB {
	return db.Select(txIdColumn, voutColumn,
		"change",
		"min(case when change >= 0 then change end) over () as "+minChange,
	).
		Table("(?) as utxoWithChange", utxoWithChange)
}

func (c *inputsQueryComposer) feeCalculatedWithChangeOutput() string {
	return fmt.Sprintf("ceil((sum(estimated_input_size) over (order by touched_at ASC, created_at ASC, tx_id ASC, vout ASC) + %d + %d) / cast(%d as float)) * %d as fee_with_change_output", c.txWithoutInputsSize, estimatedChangeOutputSize, c.feeUnit.Bytes, c.feeUnit.Satoshis)
}

func (c *inputsQueryComposer) feeCalculatedWithoutChangeOutput() string {
	return fmt.Sprintf("ceil((sum(estimated_input_size) over (order by touched_at ASC, created_at ASC, tx_id ASC, vout ASC) + %d) / cast(%d as float)) * %d as fee_no_change_output", c.txWithoutInputsSize, c.feeUnit.Bytes, c.feeUnit.Satoshis)
}

func (c *inputsQueryComposer) remainingValue() string {
	return fmt.Sprintf("sum(satoshis) over (order by touched_at ASC, created_at ASC, tx_id ASC, vout ASC) - %d as remaining_value", c.outputsTotalValue)
}
