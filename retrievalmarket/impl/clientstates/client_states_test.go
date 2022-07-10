package clientstates_test

import (
	"context"
	"errors"
	"math/rand"
	"testing"

	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/filecoin-project/go-address"
	datatransfer "github.com/filecoin-project/go-data-transfer"
	"github.com/filecoin-project/go-statemachine/fsm"
	fsmtest "github.com/filecoin-project/go-statemachine/fsm/testutil"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/abi/big"
	"github.com/filecoin-project/specs-actors/actors/builtin/paych"

	"github.com/filecoin-project/go-fil-markets/retrievalmarket"
	rm "github.com/filecoin-project/go-fil-markets/retrievalmarket"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket/impl/clientstates"
	"github.com/filecoin-project/go-fil-markets/retrievalmarket/impl/testnodes"
	testnet "github.com/filecoin-project/go-fil-markets/shared_testutil"
)

type consumeBlockResponse struct {
	size uint64
	done bool
	err  error
}

type fakeEnvironment struct {
	node                         retrievalmarket.RetrievalClientNode
	OpenDataTransferError        error
	SendDataTransferVoucherError error
	CloseDataTransferError       error
}

func (e *fakeEnvironment) Node() retrievalmarket.RetrievalClientNode {
	return e.node
}

func (e *fakeEnvironment) OpenDataTransfer(ctx context.Context, to peer.ID, proposal *rm.DealProposal) (datatransfer.ChannelID, error) {
	return datatransfer.ChannelID{ID: datatransfer.TransferID(rand.Uint64()), Responder: to, Initiator: testnet.GeneratePeers(1)[0]}, e.OpenDataTransferError
}

func (e *fakeEnvironment) SendDataTransferVoucher(_ context.Context, _ datatransfer.ChannelID, _ *rm.DealPayment) error {
	return e.SendDataTransferVoucherError
}

func (e *fakeEnvironment) CloseDataTransfer(_ context.Context, _ datatransfer.ChannelID) error {
	return e.CloseDataTransferError
}

func TestProposeDeal(t *testing.T) {
	ctx := context.Background()
	node := testnodes.NewTestRetrievalClientNode(testnodes.TestRetrievalClientNodeParams{})
	eventMachine, err := fsm.NewEventProcessor(retrievalmarket.ClientDealState{}, "Status", clientstates.ClientEvents)
	require.NoError(t, err)
	runProposeDeal := func(t *testing.T, openError error, dealState *retrievalmarket.ClientDealState) {
		environment := &fakeEnvironment{node, openError, nil, nil}
		fsmCtx := fsmtest.NewTestContext(ctx, eventMachine)
		err := clientstates.ProposeDeal(fsmCtx, environment, *dealState)
		require.NoError(t, err)
		fsmCtx.ReplayEvents(t, dealState)
	}

	t.Run("it works", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusNew)
		var openError error = nil
		runProposeDeal(t, openError, dealState)
		require.Empty(t, dealState.Message)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusWaitForAcceptance)
		require.Equal(t, dealState.ChannelID.Responder, dealState.Sender)
	})

	t.Run("data transfer eror", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusNew)
		openError := errors.New("something went wrong")
		runProposeDeal(t, openError, dealState)
		require.NotEmpty(t, dealState.Message)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusErrored)
	})
}
func TestSetupPaymentChannel(t *testing.T) {
	ctx := context.Background()
	expectedPayCh := address.TestAddress2
	eventMachine, err := fsm.NewEventProcessor(retrievalmarket.ClientDealState{}, "Status", clientstates.ClientEvents)
	require.NoError(t, err)
	runSetupPaymentChannel := func(t *testing.T,
		params testnodes.TestRetrievalClientNodeParams,
		dealState *retrievalmarket.ClientDealState) {
		node := testnodes.NewTestRetrievalClientNode(params)
		environment := &fakeEnvironment{node, nil, nil, nil}
		fsmCtx := fsmtest.NewTestContext(ctx, eventMachine)
		err := clientstates.SetupPaymentChannelStart(fsmCtx, environment, *dealState)
		require.NoError(t, err)
		fsmCtx.ReplayEvents(t, dealState)
	}

	t.Run("payment channel create initiated", func(t *testing.T) {
		envParams := testnodes.TestRetrievalClientNodeParams{
			PayCh:          address.Undef,
			CreatePaychCID: testnet.GenerateCids(1)[0],
		}
		dealState := makeDealState(retrievalmarket.DealStatusAccepted)
		runSetupPaymentChannel(t, envParams, dealState)
		assert.Empty(t, dealState.Message)
		assert.Equal(t, dealState.Status, retrievalmarket.DealStatusPaymentChannelCreating)
	})

	t.Run("payment channel needs funds added", func(t *testing.T) {
		envParams := testnodes.TestRetrievalClientNodeParams{
			AddFundsOnly:   true,
			PayCh:          expectedPayCh,
			CreatePaychCID: testnet.GenerateCids(1)[0],
		}
		dealState := makeDealState(retrievalmarket.DealStatusAccepted)
		runSetupPaymentChannel(t, envParams, dealState)
		require.Empty(t, dealState.Message)
		require.Equal(t, retrievalmarket.DealStatusPaymentChannelAddingFunds, dealState.Status)
		require.Equal(t, expectedPayCh, dealState.PaymentInfo.PayCh)
	})

	t.Run("when create payment channel fails", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusAccepted)
		envParams := testnodes.TestRetrievalClientNodeParams{
			PayCh:    address.Undef,
			PayChErr: errors.New("Something went wrong"),
		}
		runSetupPaymentChannel(t, envParams, dealState)
		require.NotEmpty(t, dealState.Message)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusFailing)
	})

}

func TestWaitForPaymentChannelCreate(t *testing.T) {
	ctx := context.Background()
	expectedPayCh := address.TestAddress2
	expectedLane := uint64(10)
	eventMachine, err := fsm.NewEventProcessor(retrievalmarket.ClientDealState{}, "Status", clientstates.ClientEvents)
	require.NoError(t, err)
	runWaitForPaychCreate := func(t *testing.T,
		params testnodes.TestRetrievalClientNodeParams,
		dealState *retrievalmarket.ClientDealState) {
		node := testnodes.NewTestRetrievalClientNode(params)
		environment := &fakeEnvironment{node, nil, nil, nil}
		fsmCtx := fsmtest.NewTestContext(ctx, eventMachine)
		err := clientstates.WaitForPaymentChannelCreate(fsmCtx, environment, *dealState)
		require.NoError(t, err)
		fsmCtx.ReplayEvents(t, dealState)
	}
	msgCID := testnet.GenerateCids(1)[0]

	t.Run("it works", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusPaymentChannelCreating)
		dealState.WaitMsgCID = &msgCID
		params := testnodes.TestRetrievalClientNodeParams{
			PayCh:          expectedPayCh,
			CreatePaychCID: msgCID,
			Lane:           expectedLane,
		}
		runWaitForPaychCreate(t, params, dealState)
		require.Empty(t, dealState.Message)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusOngoing)
		require.Equal(t, expectedLane, dealState.PaymentInfo.Lane)
		require.Equal(t, expectedPayCh, dealState.PaymentInfo.PayCh)
	})
	t.Run("if Wait fails", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusPaymentChannelCreating)
		dealState.WaitMsgCID = &msgCID
		params := testnodes.TestRetrievalClientNodeParams{
			PayCh:              expectedPayCh,
			CreatePaychCID:     msgCID,
			WaitForChCreateErr: errors.New("boom"),
		}
		runWaitForPaychCreate(t, params, dealState)
		require.Contains(t, dealState.Message, "boom")
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusFailing)
	})

	t.Run("if AllocateLane fails", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusPaymentChannelCreating)
		dealState.WaitMsgCID = &msgCID
		params := testnodes.TestRetrievalClientNodeParams{
			PayCh:          expectedPayCh,
			CreatePaychCID: msgCID,
			LaneError:      errors.New("boom"),
		}
		runWaitForPaychCreate(t, params, dealState)
		require.Contains(t, dealState.Message, "boom")
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusFailing)
	})
}

func TestWaitForPaymentChannelAddFunds(t *testing.T) {
	ctx := context.Background()
	expectedPayCh := address.TestAddress2
	expectedLane := uint64(99)
	eventMachine, err := fsm.NewEventProcessor(retrievalmarket.ClientDealState{}, "Status", clientstates.ClientEvents)
	require.NoError(t, err)
	runWaitForPaychAddFunds := func(t *testing.T,
		params testnodes.TestRetrievalClientNodeParams,
		dealState *retrievalmarket.ClientDealState) {
		node := testnodes.NewTestRetrievalClientNode(params)
		environment := &fakeEnvironment{node, nil, nil, nil}
		fsmCtx := fsmtest.NewTestContext(ctx, eventMachine)
		err := clientstates.WaitForPaymentChannelAddFunds(fsmCtx, environment, *dealState)
		require.NoError(t, err)
		fsmCtx.ReplayEvents(t, dealState)
	}
	msgCID := testnet.GenerateCids(1)[0]

	t.Run("it works", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusPaymentChannelAddingFunds)
		dealState.PaymentInfo.PayCh = expectedPayCh
		dealState.WaitMsgCID = &msgCID

		params := testnodes.TestRetrievalClientNodeParams{
			AddFundsOnly: true,
			PayCh:        expectedPayCh,
			AddFundsCID:  msgCID,
			Lane:         expectedLane,
		}
		runWaitForPaychAddFunds(t, params, dealState)
		require.Empty(t, dealState.Message)
		assert.Equal(t, retrievalmarket.DealStatusOngoing, dealState.Status)
		assert.Equal(t, expectedLane, dealState.PaymentInfo.Lane)
		assert.Equal(t, expectedPayCh, dealState.PaymentInfo.PayCh)
	})
	t.Run("if Wait fails", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusPaymentChannelAddingFunds)
		dealState.WaitMsgCID = &msgCID
		params := testnodes.TestRetrievalClientNodeParams{
			AddFundsOnly:       true,
			PayCh:              expectedPayCh,
			AddFundsCID:        msgCID,
			WaitForAddFundsErr: errors.New("boom"),
			Lane:               expectedLane,
		}
		runWaitForPaychAddFunds(t, params, dealState)
		assert.Contains(t, dealState.Message, "boom")
		assert.Equal(t, dealState.Status, retrievalmarket.DealStatusFailing)
		assert.Equal(t, uint64(0), dealState.PaymentInfo.Lane)
	})
	t.Run("if AllocateLane fails", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusPaymentChannelAddingFunds)
		dealState.WaitMsgCID = &msgCID
		params := testnodes.TestRetrievalClientNodeParams{
			AddFundsOnly: true,
			PayCh:        expectedPayCh,
			AddFundsCID:  msgCID,
			LaneError:    errors.New("boom"),
			Lane:         expectedLane,
		}
		runWaitForPaychAddFunds(t, params, dealState)
		assert.Contains(t, dealState.Message, "boom")
		assert.Equal(t, dealState.Status, retrievalmarket.DealStatusFailing)
		assert.Equal(t, uint64(0), dealState.PaymentInfo.Lane)
	})
}

func TestOngoing(t *testing.T) {
	ctx := context.Background()
	eventMachine, err := fsm.NewEventProcessor(retrievalmarket.ClientDealState{}, "Status", clientstates.ClientEvents)
	require.NoError(t, err)
	runOngoing := func(t *testing.T,
		dealState *retrievalmarket.ClientDealState) {
		node := testnodes.NewTestRetrievalClientNode(testnodes.TestRetrievalClientNodeParams{})
		environment := &fakeEnvironment{node, nil, nil, nil}
		fsmCtx := fsmtest.NewTestContext(ctx, eventMachine)
		err := clientstates.Ongoing(fsmCtx, environment, *dealState)
		require.NoError(t, err)
		fsmCtx.ReplayEvents(t, dealState)
	}

	t.Run("it works - no change", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusOngoing)
		dealState.PaymentRequested = big.Zero()
		runOngoing(t, dealState)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusOngoing)
	})

	t.Run("it works - payment requested", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusOngoing)
		runOngoing(t, dealState)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusFundsNeeded)
	})

	t.Run("it works - last payment requested", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusOngoing)
		dealState.LastPaymentRequested = true
		runOngoing(t, dealState)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusFundsNeededLastPayment)
	})
}

func TestProcessPaymentRequested(t *testing.T) {
	ctx := context.Background()
	eventMachine, err := fsm.NewEventProcessor(retrievalmarket.ClientDealState{}, "Status", clientstates.ClientEvents)
	require.NoError(t, err)
	runProcessPaymentRequested := func(t *testing.T,
		dealState *retrievalmarket.ClientDealState) {
		node := testnodes.NewTestRetrievalClientNode(testnodes.TestRetrievalClientNodeParams{})
		environment := &fakeEnvironment{node, nil, nil, nil}
		fsmCtx := fsmtest.NewTestContext(ctx, eventMachine)
		err := clientstates.ProcessPaymentRequested(fsmCtx, environment, *dealState)
		require.NoError(t, err)
		fsmCtx.ReplayEvents(t, dealState)
	}

	t.Run("it works - to send funds", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusFundsNeeded)
		runProcessPaymentRequested(t, dealState)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusSendFunds)
	})

	t.Run("it works - to send funds", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusFundsNeededLastPayment)
		dealState.TotalReceived = defaultBytesPaidFor + 500
		dealState.AllBlocksReceived = true
		runProcessPaymentRequested(t, dealState)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusSendFundsLastPayment)
	})

	t.Run("no change", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusFundsNeeded)
		dealState.BytesPaidFor = defaultBytesPaidFor + 500
		runProcessPaymentRequested(t, dealState)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusFundsNeeded)
	})
}

func TestSendFunds(t *testing.T) {
	ctx := context.Background()
	eventMachine, err := fsm.NewEventProcessor(retrievalmarket.ClientDealState{}, "Status", clientstates.ClientEvents)
	require.NoError(t, err)
	runSendFunds := func(t *testing.T,
		sendDataTransferVoucherError error,
		nodeParams testnodes.TestRetrievalClientNodeParams,
		dealState *retrievalmarket.ClientDealState) {
		node := testnodes.NewTestRetrievalClientNode(nodeParams)
		environment := &fakeEnvironment{node, nil, sendDataTransferVoucherError, nil}
		fsmCtx := fsmtest.NewTestContext(ctx, eventMachine)
		err := clientstates.SendFunds(fsmCtx, environment, *dealState)
		require.NoError(t, err)
		fsmCtx.ReplayEvents(t, dealState)
	}

	testVoucher := &paych.SignedVoucher{}

	t.Run("it works", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusSendFunds)
		var sendVoucherError error = nil
		nodeParams := testnodes.TestRetrievalClientNodeParams{
			Voucher: testVoucher,
		}
		runSendFunds(t, sendVoucherError, nodeParams, dealState)
		require.Empty(t, dealState.Message)
		require.Equal(t, dealState.PaymentRequested, abi.NewTokenAmount(0))
		require.Equal(t, dealState.FundsSpent, big.Add(defaultFundsSpent, defaultPaymentRequested))
		require.Equal(t, dealState.BytesPaidFor, defaultTotalReceived)
		require.Equal(t, dealState.CurrentInterval, defaultCurrentInterval+defaultIntervalIncrease)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusOngoing)
	})

	t.Run("last payment", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusSendFundsLastPayment)
		var sendVoucherError error = nil
		nodeParams := testnodes.TestRetrievalClientNodeParams{
			Voucher: testVoucher,
		}
		runSendFunds(t, sendVoucherError, nodeParams, dealState)
		require.Empty(t, dealState.Message)
		require.Equal(t, dealState.PaymentRequested, abi.NewTokenAmount(0))
		require.Equal(t, dealState.FundsSpent, big.Add(defaultFundsSpent, defaultPaymentRequested))
		require.Equal(t, dealState.BytesPaidFor, defaultTotalReceived)
		require.Equal(t, dealState.CurrentInterval, defaultCurrentInterval+defaultIntervalIncrease)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusFinalizing)
	})

	t.Run("not enough funds left", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusSendFunds)
		dealState.FundsSpent = defaultTotalFunds
		var sendVoucherError error = nil
		nodeParams := testnodes.TestRetrievalClientNodeParams{
			Voucher: testVoucher,
		}
		runSendFunds(t, sendVoucherError, nodeParams, dealState)
		require.NotEmpty(t, dealState.Message)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusFailing)
	})

	t.Run("more bytes since last payment than interval works, can charge more", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusSendFunds)
		dealState.BytesPaidFor = defaultBytesPaidFor - 500
		largerPaymentRequested := abi.NewTokenAmount(750000)
		dealState.PaymentRequested = largerPaymentRequested
		var sendVoucherError error = nil
		nodeParams := testnodes.TestRetrievalClientNodeParams{
			Voucher: testVoucher,
		}
		runSendFunds(t, sendVoucherError, nodeParams, dealState)
		require.Empty(t, dealState.Message)
		require.Equal(t, dealState.PaymentRequested, abi.NewTokenAmount(0))
		require.Equal(t, dealState.FundsSpent, big.Add(defaultFundsSpent, largerPaymentRequested))
		require.Equal(t, dealState.BytesPaidFor, defaultTotalReceived)
		require.Equal(t, dealState.CurrentInterval, defaultCurrentInterval+defaultIntervalIncrease)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusOngoing)
	})

	t.Run("too much payment requested", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusSendFunds)
		dealState.PaymentRequested = abi.NewTokenAmount(750000)
		var sendVoucherError error = nil
		nodeParams := testnodes.TestRetrievalClientNodeParams{
			Voucher: testVoucher,
		}
		runSendFunds(t, sendVoucherError, nodeParams, dealState)
		require.NotEmpty(t, dealState.Message)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusFailing)
	})

	t.Run("too little payment requested works but records correctly", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusSendFunds)
		smallerPaymentRequested := abi.NewTokenAmount(250000)
		dealState.PaymentRequested = smallerPaymentRequested
		var sendVoucherError error = nil
		nodeParams := testnodes.TestRetrievalClientNodeParams{
			Voucher: testVoucher,
		}
		runSendFunds(t, sendVoucherError, nodeParams, dealState)
		require.Empty(t, dealState.Message)
		require.Equal(t, dealState.PaymentRequested, abi.NewTokenAmount(0))
		require.Equal(t, dealState.FundsSpent, big.Add(defaultFundsSpent, smallerPaymentRequested))
		// only records change for those bytes paid for
		require.Equal(t, dealState.BytesPaidFor, defaultBytesPaidFor+500)
		// no interval increase
		require.Equal(t, dealState.CurrentInterval, defaultCurrentInterval)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusOngoing)
	})

	t.Run("voucher create fails", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusSendFunds)
		var sendVoucherError error = nil
		nodeParams := testnodes.TestRetrievalClientNodeParams{
			VoucherError: errors.New("Something Went Wrong"),
		}
		runSendFunds(t, sendVoucherError, nodeParams, dealState)
		require.NotEmpty(t, dealState.Message)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusFailing)
	})

	t.Run("unable to send payment", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusSendFunds)
		sendVoucherError := errors.New("something went wrong")
		nodeParams := testnodes.TestRetrievalClientNodeParams{
			Voucher: testVoucher,
		}
		runSendFunds(t, sendVoucherError, nodeParams, dealState)
		require.NotEmpty(t, dealState.Message)
		require.Equal(t, dealState.Status, retrievalmarket.DealStatusErrored)
	})
}

func TestCancelDeal(t *testing.T) {
	ctx := context.Background()
	eventMachine, err := fsm.NewEventProcessor(retrievalmarket.ClientDealState{}, "Status", clientstates.ClientEvents)
	require.NoError(t, err)
	runCancelDeal := func(t *testing.T,
		closeError error,
		dealState *retrievalmarket.ClientDealState) {
		node := testnodes.NewTestRetrievalClientNode(testnodes.TestRetrievalClientNodeParams{})
		environment := &fakeEnvironment{node, nil, nil, closeError}
		fsmCtx := fsmtest.NewTestContext(ctx, eventMachine)
		err := clientstates.CancelDeal(fsmCtx, environment, *dealState)
		require.NoError(t, err)
		fsmCtx.ReplayEvents(t, dealState)
	}

	t.Run("it works", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusFailing)
		dealState.Message = "Previous error"
		runCancelDeal(t, nil, dealState)
		require.Equal(t, "Previous error", dealState.Message)
		require.Equal(t, retrievalmarket.DealStatusErrored, dealState.Status)
	})

	t.Run("error closing stream", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusFailing)
		dealState.Message = "Previous error"
		runCancelDeal(t, errors.New("something went wrong"), dealState)
		require.NotEqual(t, "Previous error", dealState.Message)
		require.NotEmpty(t, dealState.Message)
		require.Equal(t, retrievalmarket.DealStatusErrored, dealState.Status)
	})
}
func TestCheckComplete(t *testing.T) {
	ctx := context.Background()
	eventMachine, err := fsm.NewEventProcessor(retrievalmarket.ClientDealState{}, "Status", clientstates.ClientEvents)
	require.NoError(t, err)
	runCheckComplete := func(t *testing.T,
		dealState *retrievalmarket.ClientDealState) {
		node := testnodes.NewTestRetrievalClientNode(testnodes.TestRetrievalClientNodeParams{})
		environment := &fakeEnvironment{node, nil, nil, nil}
		fsmCtx := fsmtest.NewTestContext(ctx, eventMachine)
		err := clientstates.CheckComplete(fsmCtx, environment, *dealState)
		require.NoError(t, err)
		fsmCtx.ReplayEvents(t, dealState)
	}

	t.Run("when all blocks received", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusCheckComplete)
		dealState.AllBlocksReceived = true
		runCheckComplete(t, dealState)
		require.Equal(t, retrievalmarket.DealStatusCompleted, dealState.Status)
	})

	t.Run("when not all blocks are received", func(t *testing.T) {
		dealState := makeDealState(retrievalmarket.DealStatusCheckComplete)
		dealState.AllBlocksReceived = false
		runCheckComplete(t, dealState)
		require.Equal(t, retrievalmarket.DealStatusErrored, dealState.Status)
		require.Equal(t, "Provider sent complete status without sending all data", dealState.Message)
	})
}

var defaultTotalFunds = abi.NewTokenAmount(4000000)
var defaultCurrentInterval = uint64(1000)
var defaultIntervalIncrease = uint64(500)
var defaultPricePerByte = abi.NewTokenAmount(500)
var defaultTotalReceived = uint64(6000)
var defaultBytesPaidFor = uint64(5000)
var defaultFundsSpent = abi.NewTokenAmount(2500000)
var defaultPaymentRequested = abi.NewTokenAmount(500000)
var defaultUnsealFundsPaid = abi.NewTokenAmount(0)

func makeDealState(status retrievalmarket.DealStatus) *retrievalmarket.ClientDealState {
	return &retrievalmarket.ClientDealState{
		TotalFunds:       defaultTotalFunds,
		MinerWallet:      address.TestAddress,
		ClientWallet:     address.TestAddress2,
		PaymentInfo:      &retrievalmarket.PaymentInfo{},
		Status:           status,
		BytesPaidFor:     defaultBytesPaidFor,
		TotalReceived:    defaultTotalReceived,
		CurrentInterval:  defaultCurrentInterval,
		FundsSpent:       defaultFundsSpent,
		UnsealFundsPaid:  defaultUnsealFundsPaid,
		PaymentRequested: defaultPaymentRequested,
		DealProposal: retrievalmarket.DealProposal{
			ID:     retrievalmarket.DealID(10),
			Params: retrievalmarket.NewParamsV0(defaultPricePerByte, 0, defaultIntervalIncrease),
		},
	}
}
