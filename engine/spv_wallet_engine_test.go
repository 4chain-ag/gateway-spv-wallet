package engine

import (
	"context"
	"database/sql"
	"testing"

	compat "github.com/bitcoin-sv/go-sdk/compat/bip32"
	"github.com/bitcoin-sv/spv-wallet/config"
	"github.com/bitcoin-sv/spv-wallet/engine/datastore"
	"github.com/bitcoin-sv/spv-wallet/engine/taskmanager"
	"github.com/bitcoin-sv/spv-wallet/engine/tester"
	"github.com/mrz1836/go-cache"
	"github.com/rafaeljusto/redigomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

// TestingClient is for testing the entire package using real/mocked services
type TestingClient struct {
	client      ClientInterface  // Local SPV Wallet Engine client for testing
	ctx         context.Context  // Current CTX
	database    datastore.Engine // Current database
	redisClient *cache.Client    // Current redis client (used for Mocking)
	redisConn   *redigomock.Conn // Current redis connection (used for Mocking)
	SQLConn     *sql.DB          // Read test client
	tablePrefix string           // Current table prefix
}

// Close will close all test services and client
func (tc *TestingClient) Close(ctx context.Context) {
	if tc.client != nil {
		/*if !tc.mocking {
			if err := dropAllTables(tc.client.Datastore(), tc.database); err != nil {
				panic(err)
			}
		}*/
		_ = tc.client.Close(ctx)
	}
	if tc.SQLConn != nil {
		_ = tc.SQLConn.Close()
	}
	if tc.redisClient != nil {
		tc.redisClient.Close()
	}
	if tc.redisConn != nil {
		_ = tc.redisConn.Close()
	}
}

// DefaultClientOpts will return a default set of client options required to load the new client
func DefaultClientOpts() []ClientOps {
	tqc := taskmanager.DefaultTaskQConfig(tester.RandomTablePrefix())
	tqc.MaxNumWorker = 2
	tqc.MaxNumFetcher = 2

	opts := make([]ClientOps, 0)
	opts = append(
		opts,
		WithTaskqConfig(tqc),
		WithSQLite(tester.SQLiteTestConfig()),
		WithCustomFeeUnit(mockFeeUnit),
		WithAppConfig(&config.AppConfig{
			TokenOverlay: &config.TokenOverlayConfig{
				URL: "http://localhost:3091",
			},
			Gateway: &config.GatewayConfig{
				URL: "http://localhost:8090",
			},
		}),
	)

	return opts
}

// CreateTestSQLiteClient will create a test client for SQLite
//
// NOTE: you need to close the client using the returned defer func
func CreateTestSQLiteClient(t *testing.T, debug, shared bool, clientOpts ...ClientOps) (context.Context, ClientInterface, func()) {
	ctx := context.Background()

	logger := zerolog.Nop()

	// Set the default options, add migrate models
	opts := DefaultClientOpts()
	opts = append(opts, WithLogger(&logger))
	opts = append(opts, clientOpts...)

	// Create the client
	client, err := NewClient(ctx, opts...)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Create a defer function
	f := func() {
		_ = client.Close(context.Background())
	}
	return ctx, client, f
}

// CreateBenchmarkSQLiteClient will create a test client for SQLite
//
// NOTE: you need to close the client using the returned defer func
func CreateBenchmarkSQLiteClient(b *testing.B, debug, shared bool, clientOpts ...ClientOps) (context.Context, ClientInterface, func()) {
	ctx := context.Background()

	logger := zerolog.Nop()

	// Set the default options, add migrate models
	opts := DefaultClientOpts()
	opts = append(opts, WithLogger(&logger))
	opts = append(opts, clientOpts...)

	// Create the client
	client, err := NewClient(ctx, opts...)
	if err != nil {
		b.Fail()
	}

	// Create a defer function
	f := func() {
		_ = client.Close(context.Background())
	}
	return ctx, client, f
}

// CloseClient is function used in the "defer()" function
func CloseClient(ctx context.Context, t *testing.T, client ClientInterface) {
	require.NoError(t, client.Close(ctx))
}

// CreateNewXPub will create a new xPub and return all the information to use the xPub
func CreateNewXPub(ctx context.Context, t *testing.T, engineClient ClientInterface,
	opts ...ModelOps,
) (*compat.ExtendedKey, *Xpub, string) {
	// Generate a key pair
	masterKey, err := compat.GenerateHDKey(compat.SecureSeedLength)
	require.NoError(t, err)
	require.NotNil(t, masterKey)

	// Get the raw string of the xPub
	var rawXPub string
	rawXPub, err = compat.GetExtendedPublicKey(masterKey)
	require.NoError(t, err)
	require.NotNil(t, masterKey)

	// Create the new xPub
	var xPub *Xpub
	xPub, err = engineClient.NewXpub(ctx, rawXPub, opts...)
	require.NoError(t, err)
	require.NotNil(t, xPub)

	return masterKey, xPub, rawXPub
}
