package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"

	"github.com/BuxOrg/bux"
	"github.com/mrz1836/go-datastore"
)

// AdminPaymailCreate is the resolver for the admin_paymail_create field.
func (r *mutationResolver) AdminPaymailCreate(ctx context.Context, xpub string, address string, publicName *string, avatar *string, metadata bux.Metadata) (*bux.PaymailAddress, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}

	opts := c.Services.Bux.DefaultModelOptions()

	if metadata != nil {
		opts = append(opts, bux.WithMetadatas(metadata))
	}

	usePublicName := ""
	if publicName != nil {
		usePublicName = *publicName
	}
	useAvatar := ""
	if avatar != nil {
		useAvatar = *avatar
	}

	var paymailAddress *bux.PaymailAddress
	paymailAddress, err = c.Services.Bux.NewPaymailAddress(ctx, xpub, address, usePublicName, useAvatar, opts...)
	if err != nil {
		return nil, err
	}

	return paymailAddress, nil
}

// AdminPaymailDelete is the resolver for the admin_paymail_delete field.
func (r *mutationResolver) AdminPaymailDelete(ctx context.Context, address string) (bool, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return false, err
	}

	opts := c.Services.Bux.DefaultModelOptions()

	// Delete a new paymail address
	err = c.Services.Bux.DeletePaymailAddress(ctx, address, opts...)
	if err != nil {
		return false, err
	}

	return true, nil
}

// AdminTransaction is the resolver for the admin_transaction field.
func (r *mutationResolver) AdminTransaction(ctx context.Context, hex string) (*bux.Transaction, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}

	opts := c.Services.Bux.DefaultModelOptions()

	var transaction *bux.Transaction
	transaction, err = c.Services.Bux.RecordRawTransaction(
		ctx, hex, opts...,
	)
	if err != nil {
		// already registered, just return the registered transaction
		if errors.Is(err, datastore.ErrDuplicateKey) {
			if transaction, err = c.Services.Bux.GetTransactionByHex(ctx, hex); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return bux.DisplayModels(transaction).(*bux.Transaction), nil
}

// AdminGetStatus is the resolver for the admin_get_status field.
func (r *queryResolver) AdminGetStatus(ctx context.Context) (*bool, error) {
	// including admin check
	_, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}

	success := true
	return &success, nil
}

// AdminGetStats is the resolver for the admin_get_stats field.
func (r *queryResolver) AdminGetStats(ctx context.Context) (*bux.AdminStats, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var accessKeys *bux.AdminStats
	accessKeys, err = c.Services.Bux.GetStats(ctx, c.Services.Bux.DefaultModelOptions()...)
	if err != nil {
		return nil, err
	}

	return accessKeys, nil
}

// AdminAccessKeysList is the resolver for the admin_access_keys_list field.
func (r *queryResolver) AdminAccessKeysList(ctx context.Context, metadata bux.Metadata, conditions map[string]interface{}, params *datastore.QueryParams) ([]*bux.AccessKey, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var accessKeys []*bux.AccessKey
	accessKeys, err = c.Services.Bux.GetAccessKeys(ctx, &metadata, &conditions, params)
	if err != nil {
		return nil, err
	}

	return accessKeys, nil
}

// AdminAccessKeysCount is the resolver for the admin_access_keys_count field.
func (r *queryResolver) AdminAccessKeysCount(ctx context.Context, metadata bux.Metadata, conditions map[string]interface{}) (*int64, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var count int64
	count, err = c.Services.Bux.GetAccessKeysCount(ctx, &metadata, &conditions)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

// AdminBlockHeadersList is the resolver for the admin_block_headers_list field.
func (r *queryResolver) AdminBlockHeadersList(ctx context.Context, metadata bux.Metadata, conditions map[string]interface{}, params *datastore.QueryParams) ([]*bux.BlockHeader, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var blockHeaders []*bux.BlockHeader
	blockHeaders, err = c.Services.Bux.GetBlockHeaders(ctx, &metadata, &conditions, params)
	if err != nil {
		return nil, err
	}

	return blockHeaders, nil
}

// AdminBlockHeadersCount is the resolver for the admin_block_headers_count field.
func (r *queryResolver) AdminBlockHeadersCount(ctx context.Context, metadata bux.Metadata, conditions map[string]interface{}) (*int64, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var count int64
	count, err = c.Services.Bux.GetBlockHeadersCount(ctx, &metadata, &conditions)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

// AdminDestinationsList is the resolver for the admin_destinations_list field.
func (r *queryResolver) AdminDestinationsList(ctx context.Context, metadata bux.Metadata, conditions map[string]interface{}, params *datastore.QueryParams) ([]*bux.Destination, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var destinations []*bux.Destination
	destinations, err = c.Services.Bux.GetDestinations(ctx, &metadata, &conditions, params)
	if err != nil {
		return nil, err
	}

	return destinations, nil
}

// AdminDestinationsCount is the resolver for the admin_destinations_count field.
func (r *queryResolver) AdminDestinationsCount(ctx context.Context, metadata bux.Metadata, conditions map[string]interface{}) (*int64, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var count int64
	count, err = c.Services.Bux.GetDestinationsCount(ctx, &metadata, &conditions)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

// AdminDraftTransactionsList is the resolver for the admin_draft_transactions_list field.
func (r *queryResolver) AdminDraftTransactionsList(ctx context.Context, metadata bux.Metadata, conditions map[string]interface{}, params *datastore.QueryParams) ([]*bux.DraftTransaction, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var draftTransactions []*bux.DraftTransaction
	draftTransactions, err = c.Services.Bux.GetDraftTransactions(ctx, &metadata, &conditions, params)
	if err != nil {
		return nil, err
	}

	return draftTransactions, nil
}

// AdminDraftTransactionsCount is the resolver for the admin_draft_transactions_count field.
func (r *queryResolver) AdminDraftTransactionsCount(ctx context.Context, metadata bux.Metadata, conditions map[string]interface{}) (*int64, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var count int64
	count, err = c.Services.Bux.GetDraftTransactionsCount(ctx, &metadata, &conditions)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

// AdminPaymailGet is the resolver for the admin_paymail_get field.
func (r *queryResolver) AdminPaymailGet(ctx context.Context, address string) (*bux.PaymailAddress, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}

	opts := c.Services.Bux.DefaultModelOptions()

	var paymailAddress *bux.PaymailAddress
	paymailAddress, err = c.Services.Bux.GetPaymailAddress(ctx, address, opts...)
	if err != nil {
		return nil, err
	}

	return paymailAddress, nil
}

// AdminPaymailGetByXpubID is the resolver for the admin_paymail_get_by_xpub_id field.
func (r *queryResolver) AdminPaymailGetByXpubID(ctx context.Context, xpubID string) ([]*bux.PaymailAddress, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var paymailAddresses []*bux.PaymailAddress
	paymailAddresses, err = c.Services.Bux.GetPaymailAddressesByXPubID(ctx, xpubID, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	return paymailAddresses, nil
}

// AdminPaymailsList is the resolver for the admin_paymails_list field.
func (r *queryResolver) AdminPaymailsList(ctx context.Context, metadata bux.Metadata, conditions map[string]interface{}, params *datastore.QueryParams) ([]*bux.PaymailAddress, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var paymailAddresses []*bux.PaymailAddress
	paymailAddresses, err = c.Services.Bux.GetPaymailAddresses(ctx, &metadata, &conditions, nil)
	if err != nil {
		return nil, err
	}

	return paymailAddresses, nil
}

// AdminPaymailsCount is the resolver for the admin_paymails_count field.
func (r *queryResolver) AdminPaymailsCount(ctx context.Context, metadata bux.Metadata, conditions map[string]interface{}) (*int64, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var count int64
	count, err = c.Services.Bux.GetPaymailAddressesCount(ctx, &metadata, &conditions)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

// AdminTransactionsList is the resolver for the admin_transactions_list field.
func (r *queryResolver) AdminTransactionsList(ctx context.Context, metadata bux.Metadata, conditions map[string]interface{}, params *datastore.QueryParams) ([]*bux.Transaction, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var transactions []*bux.Transaction
	transactions, err = c.Services.Bux.GetTransactions(ctx, &metadata, &conditions, params)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}

// AdminTransactionsCount is the resolver for the admin_transactions_count field.
func (r *queryResolver) AdminTransactionsCount(ctx context.Context, metadata bux.Metadata, conditions map[string]interface{}) (*int64, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var count int64
	count, err = c.Services.Bux.GetTransactionsCount(ctx, &metadata, &conditions)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

// AdminUtxosList is the resolver for the admin_utxos_list field.
func (r *queryResolver) AdminUtxosList(ctx context.Context, metadata bux.Metadata, conditions map[string]interface{}, params *datastore.QueryParams) ([]*bux.Utxo, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var utxos []*bux.Utxo
	utxos, err = c.Services.Bux.GetUtxos(ctx, &metadata, &conditions, params)
	if err != nil {
		return nil, err
	}

	return utxos, nil
}

// AdminUtxosCount is the resolver for the admin_utxos_count field.
func (r *queryResolver) AdminUtxosCount(ctx context.Context, metadata bux.Metadata, conditions map[string]interface{}) (*int64, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var count int64
	count, err = c.Services.Bux.GetUtxosCount(ctx, &metadata, &conditions)
	if err != nil {
		return nil, err
	}

	return &count, nil
}

// AdminXpubsList is the resolver for the admin_xpubs_list field.
func (r *queryResolver) AdminXpubsList(ctx context.Context, metadata bux.Metadata, conditions map[string]interface{}, params *datastore.QueryParams) ([]*bux.Xpub, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}
	var xpubs []*bux.Xpub
	xpubs, err = c.Services.Bux.GetXPubs(ctx, &metadata, &conditions, params)
	if err != nil {
		return nil, err
	}

	return xpubs, nil
}

// AdminXpubsCount is the resolver for the admin_xpubs_count field.
func (r *queryResolver) AdminXpubsCount(ctx context.Context, metadata bux.Metadata, conditions map[string]interface{}) (*int64, error) {
	// including admin check
	c, err := GetConfigFromContextAdmin(ctx)
	if err != nil {
		return nil, err
	}

	var count int64
	count, err = c.Services.Bux.GetXPubsCount(ctx, &metadata, &conditions)
	if err != nil {
		return nil, err
	}

	return &count, nil
}
