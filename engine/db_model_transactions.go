package engine

import (
	"context"

	"github.com/bitcoin-sv/spv-wallet/engine/datastore"
	"github.com/bitcoin-sv/spv-wallet/engine/notifications"
	"github.com/bitcoin-sv/spv-wallet/engine/spverrors"
	"github.com/bitcoin-sv/spv-wallet/models"
)

// GetModelTableName will get the db table name of the current model
func (m *Transaction) GetModelTableName() string {
	return tableTransactions
}

// Save will save the model into the Datastore
func (m *Transaction) Save(ctx context.Context) (err error) {
	// Prepare the metadata
	if len(m.Metadata) > 0 {
		// set the metadata to be xpub specific, but only if we have a valid xpub ID
		if m.XPubID != "" {
			// was metadata set via opts ?
			if m.XpubMetadata == nil {
				m.XpubMetadata = make(XpubMetadata)
			}
			if _, ok := m.XpubMetadata[m.XPubID]; !ok {
				m.XpubMetadata[m.XPubID] = make(Metadata)
			}
			for key, value := range m.Metadata {
				m.XpubMetadata[m.XPubID][key] = value
			}
		} else {
			m.Client().Logger().Debug().
				Str("txID", m.ID).
				Msg("xPub id is missing from transaction, cannot store metadata")
		}
	}

	return Save(ctx, m)
}

// BeforeCreating will fire before the model is being inserted into the Datastore
func (m *Transaction) BeforeCreating(_ context.Context) error {
	if m.beforeCreateCalled {
		m.Client().Logger().Debug().
			Str("txID", m.ID).
			Msgf("skipping: %s BeforeCreating hook, because already called", m.Name())
		return nil
	}

	m.Client().Logger().Debug().
		Str("txID", m.ID).
		Msgf("starting: %s BeforeCreating hook...", m.Name())

	// Test for required field(s)
	if len(m.Hex) == 0 {
		return ErrMissingFieldHex
	}

	// Set the xPubID
	m.setXPubID()

	// Set the ID - will also parse and verify the tx
	err := m.setID()
	if err != nil {
		return err
	}

	m.Client().Logger().Debug().
		Str("txID", m.ID).
		Msgf("end: %s BeforeCreating hook", m.Name())
	m.beforeCreateCalled = true
	return nil
}

// AfterCreated will fire after the model is created in the Datastore
func (m *Transaction) AfterCreated(ctx context.Context) error {
	m.Client().Logger().Debug().
		Str("txID", m.ID).
		Msgf("starting: %s AfterCreated hook...", m.Name())

	// Pre-build the options
	opts := m.GetOptions(false)

	// update the xpub balances
	for xPubID, balance := range m.XpubOutputValue {
		// todo: run this in a go routine? (move this into a function on the xpub model?)
		xPub, err := getXpubWithCache(ctx, m.Client(), "", xPubID, opts...)
		if err != nil {
			return err
		} else if xPub == nil {
			return spverrors.ErrMissingFieldXpub
		}
		if err = xPub.incrementBalance(ctx, balance); err != nil {
			return err
		}
	}

	// Update the draft transaction, process broadcasting
	// todo: go routine (however it's not working, panic in save for missing datastore)
	if m.draftTransaction != nil {
		m.draftTransaction.Status = DraftStatusComplete
		m.draftTransaction.FinalTxID = m.ID
		if err := m.draftTransaction.Save(ctx); err != nil {
			return err
		}
	}

	m.notify()

	m.Client().Logger().Debug().
		Str("txID", m.ID).
		Msgf("end: %s AfterCreated hook...", m.Name())
	return nil
}

// AfterUpdated will fire after the model is updated in the Datastore
func (m *Transaction) AfterUpdated(_ context.Context) error {
	m.Client().Logger().Debug().
		Str("txID", m.ID).
		Msgf("starting: %s AfterUpdated hook...", m.Name())

	m.notify()

	m.Client().Logger().Debug().
		Str("txID", m.ID).
		Msgf("end: %s AfterUpdated hook", m.Name())
	return nil
}

// AfterDeleted will fire after the model is deleted in the Datastore
func (m *Transaction) AfterDeleted(_ context.Context) error {
	m.Client().Logger().Debug().Msgf("starting: %s AfterDeleted hook...", m.Name())

	m.Client().Logger().Debug().Msgf("end: %s AfterDeleted hook", m.Name())
	return nil
}

// ChildModels will get any related sub models
func (m *Transaction) ChildModels() (childModels []ModelInterface) {
	// Add the UTXOs if found
	for index := range m.utxos {
		childModels = append(childModels, &m.utxos[index])
	}

	return
}

// PostMigrate runs right after the model is migrated
func (m *Transaction) PostMigrate(client datastore.ClientInterface) error {
	tableName := client.GetTableName(tableTransactions)
	if client.Engine() == datastore.PostgreSQL {
		if err := m.migratePostgreSQL(client, tableName); err != nil {
			return err
		}
	}

	err := client.IndexMetadata(tableName, xPubMetadataField)
	return spverrors.Wrapf(err, "failed to index metadata column on model %s", m.GetModelName())
}

// migratePostgreSQL is specific migration SQL for Postgresql
func (m *Transaction) migratePostgreSQL(client datastore.ClientInterface, tableName string) error {
	tx := client.Execute(`CREATE INDEX IF NOT EXISTS idx_` + tableName + `_xpub_in_ids ON ` +
		tableName + ` USING gin (xpub_in_ids jsonb_ops)`)
	if tx.Error != nil {
		return tx.Error
	}

	if tx = client.Execute(`CREATE INDEX IF NOT EXISTS idx_` + tableName + `_xpub_out_ids ON ` +
		tableName + ` USING gin (xpub_out_ids jsonb_ops)`); tx.Error != nil {
		return tx.Error
	}

	return nil
}

func (m *Transaction) notify() {
	if n := m.Client().Notifications(); n != nil {
		notifications.Notify(n, &models.TransactionEvent{
			UserEvent: models.UserEvent{
				XPubID: m.XPubID,
			},
			TransactionID:   m.ID,
			Status:          string(m.TxStatus),
			XpubOutputValue: m.XpubOutputValue,
		})
	}
}
