package testabilities

import (
	"math"
	"testing"

	sdk "github.com/bitcoin-sv/go-sdk/transaction"
	testpaymail "github.com/bitcoin-sv/spv-wallet/engine/paymail/testabilities"
	"github.com/bitcoin-sv/spv-wallet/engine/tester/fixtures/txtestability"
	"github.com/bitcoin-sv/spv-wallet/engine/v2/transaction/outlines"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TransactionOutlineAssertion interface {
	Created(transaction *outlines.Transaction) CreatedTransactionOutlineAssertion
	ExternalPaymailHost() testpaymail.PaymailExternalAssertions
}

type CreatedTransactionOutlineAssertion interface {
	WithNoError(err error) SuccessfullyCreatedTransactionOutlineAssertion
	WithError(err error) ErrorCreationTransactionOutlineAssertion
}

type ErrorCreationTransactionOutlineAssertion interface {
	ThatIs(expectedError error)
}

type SuccessfullyCreatedTransactionOutlineAssertion interface {
	WithParseableBEEFHex() WithParseableBEEFTransactionOutlineAssertion
	WithParseableRawHex() WithParseableBEEFTransactionOutlineAssertion
}

type WithParseableBEEFTransactionOutlineAssertion interface {
	// IsWithoutTimeLock asserts that transaction is final,
	// what means that it has all inputs with nSequence max
	// and nLockTime in past (actually we're assuming it's 0)
	IsWithoutTimeLock() WithParseableBEEFTransactionOutlineAssertion
	HasInputs(count int) WithParseableBEEFTransactionOutlineAssertion
	Input(index int) InputAssertion
	HasOutputs(count int) WithParseableBEEFTransactionOutlineAssertion
	Output(index uint32) OutputAssertion
}

func Then(t testing.TB, fixture TransactionOutlineFixture) TransactionOutlineAssertion {
	// TODO: move some methods from this package to txtestability.TransactionAssertion, to make it reusable
	return &assertion{
		t:                 t,
		require:           require.New(t),
		assert:            assert.New(t),
		paymailAssertions: testpaymail.Then(t, fixture.ExternalRecipientHost().MockedPaymailClient()),
		txFixture:         txtestability.Given(t),
	}
}

type assertion struct {
	t                 testing.TB
	require           *require.Assertions
	assert            *assert.Assertions
	txOutline         *outlines.Transaction
	tx                *sdk.Transaction
	err               error
	paymailAssertions testpaymail.PaymailExternalAssertions
	txFixture         txtestability.TransactionsFixtures
}

func (a *assertion) Created(transaction *outlines.Transaction) CreatedTransactionOutlineAssertion {
	a.t.Helper()
	a.txOutline = transaction
	return a
}

func (a *assertion) WithError(err error) ErrorCreationTransactionOutlineAssertion {
	a.t.Helper()
	a.assert.Nil(a.txOutline)
	a.assert.Error(err)
	a.err = err
	return a
}

func (a *assertion) ThatIs(expectedError error) {
	a.t.Helper()
	a.assert.ErrorIs(a.err, expectedError)
}

// WithNoError checks if there was no error and result is not nil.
func (a *assertion) WithNoError(err error) SuccessfullyCreatedTransactionOutlineAssertion {
	a.t.Helper()
	a.require.NoError(err, "Creation of transaction outline has finished with error")
	a.require.NotNil(a.txOutline, "The result is nil although there was no error")
	return a
}

func (a *assertion) WithParseableBEEFHex() WithParseableBEEFTransactionOutlineAssertion {
	a.t.Helper()
	a.t.Logf("Hex: %s", a.txOutline.Hex)

	var err error
	a.tx, err = a.txOutline.Hex.ToBEEFTransaction()
	a.require.NoErrorf(err, "Invalid BEEF hex: %s", a.txOutline.Hex)
	return a
}

func (a *assertion) IsWithoutTimeLock() WithParseableBEEFTransactionOutlineAssertion {
	a.assert.Zero(a.tx.LockTime, "Unexpected lock time on transaction")

	lo.ForEach(a.tx.Inputs, func(input *sdk.TransactionInput, index int) {
		a.assert.EqualValuesf(math.MaxUint32, input.SequenceNumber, "Unexpected sequence number value on input %d", index)
	})

	return a
}

func (a *assertion) WithParseableRawHex() WithParseableBEEFTransactionOutlineAssertion {
	a.t.Helper()
	a.t.Logf("Hex: %s", a.txOutline.Hex)

	var err error
	a.tx, err = a.txOutline.Hex.ToRawTransaction()
	a.require.NoErrorf(err, "Invalid Raw hex: %s", a.txOutline.Hex)
	return a
}

func (a *assertion) HasInputs(count int) WithParseableBEEFTransactionOutlineAssertion {
	a.t.Helper()
	a.require.Lenf(a.tx.Inputs, count, "Number of Transaction Inputs")
	return a
}

func (a *assertion) Input(index int) InputAssertion {
	a.t.Helper()
	a.require.Greater(len(a.tx.Inputs), index, "Transaction Inputs doesn't have input %d", index)

	return &txInputAssertion{
		parent:     a,
		t:          a.t,
		assert:     a.assert,
		require:    a.require,
		input:      a.tx.Inputs[index],
		annotation: a.txOutline.Annotations.Inputs[index],
		index:      index,
	}
}

func (a *assertion) HasOutputs(count int) WithParseableBEEFTransactionOutlineAssertion {
	a.t.Helper()
	a.require.Lenf(a.tx.Outputs, count, "Number of Transaction Outputs")
	a.require.Lenf(a.txOutline.Annotations.Outputs, count, "Number of Output Annotations")
	return a
}

func (a *assertion) Output(index uint32) OutputAssertion {
	a.t.Helper()
	a.require.Greater(uint32(len(a.tx.Outputs)), index, "Transaction Outputs doesn't have output %d", index)

	return &txOutputAssertion{
		parent:     a,
		t:          a.t,
		assert:     a.assert,
		require:    a.require,
		txout:      a.tx.Outputs[index],
		annotation: a.txOutline.Annotations.Outputs[index],
		index:      index,
		txFixture:  a.txFixture,
	}
}

func (a *assertion) ExternalPaymailHost() testpaymail.PaymailExternalAssertions {
	return a.paymailAssertions
}
