package dtutils_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ipfs/go-cid"
	"github.com/ipfs/go-datastore"
	"github.com/ipfs/go-graphsync/cidset"
	"github.com/ipld/go-ipld-prime"
	peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/require"

	datatransfer "github.com/filecoin-project/go-data-transfer"
	"github.com/filecoin-project/go-multistore"
	"github.com/filecoin-project/go-statemachine/fsm"

	"github.com/filecoin-project/go-fil-markets/shared_testutil"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-fil-markets/storagemarket/impl/dtutils"
	"github.com/filecoin-project/go-fil-markets/storagemarket/impl/requestvalidation"
)

func TestProviderDataTransferSubscriber(t *testing.T) {
	expectedProposalCID := shared_testutil.GenerateCids(1)[0]
	tests := map[string]struct {
		code          datatransfer.EventCode
		message       string
		status        datatransfer.Status
		called        bool
		voucher       datatransfer.Voucher
		expectedID    interface{}
		expectedEvent fsm.EventName
		expectedArgs  []interface{}
	}{
		"not a storage voucher": {
			called:  false,
			voucher: nil,
		},
		"open event": {
			code:   datatransfer.Open,
			status: datatransfer.Requested,
			called: true,
			voucher: &requestvalidation.StorageDataTransferVoucher{
				Proposal: expectedProposalCID,
			},
			expectedID:    expectedProposalCID,
			expectedEvent: storagemarket.ProviderEventDataTransferInitiated,
		},
		"completion status": {
			code:   datatransfer.Complete,
			status: datatransfer.Completed,
			called: true,
			voucher: &requestvalidation.StorageDataTransferVoucher{
				Proposal: expectedProposalCID,
			},
			expectedID:    expectedProposalCID,
			expectedEvent: storagemarket.ProviderEventDataTransferCompleted,
		},
		"data received": {
			code:   datatransfer.DataReceived,
			status: datatransfer.Ongoing,
			called: false,
			voucher: &requestvalidation.StorageDataTransferVoucher{
				Proposal: expectedProposalCID,
			},
			expectedID: expectedProposalCID,
		},
		"error event": {
			code:    datatransfer.Error,
			message: "something went wrong",
			status:  datatransfer.Failed,
			called:  true,
			voucher: &requestvalidation.StorageDataTransferVoucher{
				Proposal: expectedProposalCID,
			},
			expectedID:    expectedProposalCID,
			expectedEvent: storagemarket.ProviderEventDataTransferFailed,
			expectedArgs:  []interface{}{errors.New("deal data transfer failed: something went wrong")},
		},
		"other event": {
			code:   datatransfer.DataSent,
			status: datatransfer.Ongoing,
			called: false,
			voucher: &requestvalidation.StorageDataTransferVoucher{
				Proposal: expectedProposalCID,
			},
		},
	}
	for test, data := range tests {
		t.Run(test, func(t *testing.T) {
			ds := datastore.NewMapDatastore()
			fdg := &fakeDealGroup{}
			subscriber := dtutils.ProviderDataTransferSubscriber(ds, fdg)
			subscriber(datatransfer.Event{Code: data.code, Message: data.message}, shared_testutil.NewTestChannel(
				shared_testutil.TestChannelParams{Vouchers: []datatransfer.Voucher{data.voucher}, Status: data.status},
			))
			if data.called {
				require.True(t, fdg.called)
				require.Equal(t, fdg.lastID, data.expectedID)
				require.Equal(t, fdg.lastEvent, data.expectedEvent)
				require.Equal(t, fdg.lastArgs, data.expectedArgs)
			} else {
				require.False(t, fdg.called)
			}
		})
	}
}

func TestDataReceivedCidsPersisted(t *testing.T) {
	ds := datastore.NewMapDatastore()
	fdg := &fakeDealGroup{}
	subscriber := dtutils.ProviderDataTransferSubscriber(ds, fdg)
	expectedProposalCID1 := shared_testutil.GenerateCids(1)[0]
	expectedProposalCID2 := shared_testutil.GenerateCids(1)[0]

	// send a DataReceived event with 5 different cids for the same proposal
	cids := shared_testutil.GenerateCids(5)
	for i := 0; i < 5; i++ {
		subscriber(datatransfer.Event{Code: datatransfer.DataReceived}, shared_testutil.NewTestChannel(
			shared_testutil.TestChannelParams{Vouchers: []datatransfer.Voucher{&requestvalidation.StorageDataTransferVoucher{
				Proposal: expectedProposalCID1,
			}}, Status: datatransfer.Ongoing,
				ReceivedCids: cids[0 : i+1]},
		))
	}

	// again, with some duplicates to test de-duplication
	subscriber(datatransfer.Event{Code: datatransfer.DataReceived}, shared_testutil.NewTestChannel(
		shared_testutil.TestChannelParams{Vouchers: []datatransfer.Voucher{&requestvalidation.StorageDataTransferVoucher{
			Proposal: expectedProposalCID1,
		}}, Status: datatransfer.Ongoing,
			ReceivedCids: []cid.Cid{cids[0], cids[1], cids[1], cids[2], cids[4], cids[3], cids[4]}},
	))

	// for another proposal
	subscriber(datatransfer.Event{Code: datatransfer.DataReceived}, shared_testutil.NewTestChannel(
		shared_testutil.TestChannelParams{Vouchers: []datatransfer.Voucher{&requestvalidation.StorageDataTransferVoucher{
			Proposal: expectedProposalCID2,
		}}, Status: datatransfer.Ongoing,
			ReceivedCids: []cid.Cid{cids[0], cids[1], cids[1]}},
	))

	// assert they are all persisted in the data store without duplication for proposal 1
	key := dtutils.ReceivedCidsKey(expectedProposalCID1)
	val, err := ds.Get(key)
	require.NoError(t, err)

	set, err := cidset.DecodeCidSet(val)
	require.NoError(t, err)

	require.Equal(t, len(cids), set.Len())

	for _, c := range cids {
		require.True(t, set.Has(c))
	}

	// assert for proposal 2
	key = dtutils.ReceivedCidsKey(expectedProposalCID2)
	val, err = ds.Get(key)
	require.NoError(t, err)

	set, err = cidset.DecodeCidSet(val)
	require.NoError(t, err)

	require.Equal(t, 2, set.Len())
	require.True(t, set.Has(cids[0]))
	require.True(t, set.Has(cids[1]))
}

func TestClientDataTransferSubscriber(t *testing.T) {
	expectedProposalCID := shared_testutil.GenerateCids(1)[0]
	tests := map[string]struct {
		code          datatransfer.EventCode
		message       string
		status        datatransfer.Status
		called        bool
		voucher       datatransfer.Voucher
		expectedID    interface{}
		expectedEvent fsm.EventName
		expectedArgs  []interface{}
	}{
		"not a storage voucher": {
			called:  false,
			voucher: nil,
		},
		"completion event": {
			code:   datatransfer.Complete,
			status: datatransfer.Completed,
			called: true,
			voucher: &requestvalidation.StorageDataTransferVoucher{
				Proposal: expectedProposalCID,
			},
			expectedID:    expectedProposalCID,
			expectedEvent: storagemarket.ClientEventDataTransferComplete,
		},
		"error event": {
			code:    datatransfer.Error,
			message: "something went wrong",
			status:  datatransfer.Failed,
			called:  true,
			voucher: &requestvalidation.StorageDataTransferVoucher{
				Proposal: expectedProposalCID,
			},
			expectedID:    expectedProposalCID,
			expectedEvent: storagemarket.ClientEventDataTransferFailed,
			expectedArgs:  []interface{}{errors.New("deal data transfer failed: something went wrong")},
		},
		"other event": {
			code:   datatransfer.DataReceived,
			status: datatransfer.Ongoing,
			called: false,
			voucher: &requestvalidation.StorageDataTransferVoucher{
				Proposal: expectedProposalCID,
			},
		},
	}
	for test, data := range tests {
		t.Run(test, func(t *testing.T) {
			fdg := &fakeDealGroup{}
			subscriber := dtutils.ClientDataTransferSubscriber(fdg)
			subscriber(datatransfer.Event{Code: data.code, Message: data.message}, shared_testutil.NewTestChannel(
				shared_testutil.TestChannelParams{Vouchers: []datatransfer.Voucher{data.voucher}, Status: data.status},
			))
			if data.called {
				require.True(t, fdg.called)
				require.Equal(t, fdg.lastID, data.expectedID)
				require.Equal(t, fdg.lastEvent, data.expectedEvent)
				require.Equal(t, fdg.lastArgs, data.expectedArgs)
			} else {
				require.False(t, fdg.called)
			}
		})
	}
}

func TestTransportConfigurer(t *testing.T) {
	expectedProposalCID := shared_testutil.GenerateCids(1)[0]
	expectedChannelID := shared_testutil.MakeTestChannelID()

	testCases := map[string]struct {
		voucher          datatransfer.Voucher
		transport        datatransfer.Transport
		returnedStore    *multistore.Store
		returnedStoreErr error
		getterCalled     bool
		useStoreCalled   bool
	}{
		"non-storage voucher": {
			voucher:      nil,
			getterCalled: false,
		},
		"non-configurable transport": {
			voucher: &requestvalidation.StorageDataTransferVoucher{
				Proposal: expectedProposalCID,
			},
			transport:    &fakeTransport{},
			getterCalled: false,
		},
		"store getter errors": {
			voucher: &requestvalidation.StorageDataTransferVoucher{
				Proposal: expectedProposalCID,
			},
			transport:        &fakeGsTransport{Transport: &fakeTransport{}},
			getterCalled:     true,
			useStoreCalled:   false,
			returnedStore:    nil,
			returnedStoreErr: errors.New("something went wrong"),
		},
		"store getter succeeds": {
			voucher: &requestvalidation.StorageDataTransferVoucher{
				Proposal: expectedProposalCID,
			},
			transport:        &fakeGsTransport{Transport: &fakeTransport{}},
			getterCalled:     true,
			useStoreCalled:   true,
			returnedStore:    &multistore.Store{},
			returnedStoreErr: nil,
		},
	}
	for testCase, data := range testCases {
		t.Run(testCase, func(t *testing.T) {
			storeGetter := &fakeStoreGetter{returnedErr: data.returnedStoreErr, returnedStore: data.returnedStore}
			transportConfigurer := dtutils.TransportConfigurer(storeGetter)
			transportConfigurer(expectedChannelID, data.voucher, data.transport)
			if data.getterCalled {
				require.True(t, storeGetter.called)
				require.Equal(t, expectedProposalCID, storeGetter.lastProposalCid)
				fgt, ok := data.transport.(*fakeGsTransport)
				require.True(t, ok)
				if data.useStoreCalled {
					require.True(t, fgt.called)
					require.Equal(t, expectedChannelID, fgt.lastChannelID)
				} else {
					require.False(t, fgt.called)
				}
			} else {
				require.False(t, storeGetter.called)
			}
		})
	}
}

type fakeDealGroup struct {
	returnedErr error
	called      bool
	lastID      interface{}
	lastEvent   fsm.EventName
	lastArgs    []interface{}
}

func (fdg *fakeDealGroup) Send(id interface{}, name fsm.EventName, args ...interface{}) (err error) {
	fdg.lastID = id
	fdg.lastEvent = name
	fdg.lastArgs = args
	fdg.called = true
	return fdg.returnedErr
}

type fakeStoreGetter struct {
	lastProposalCid cid.Cid
	returnedErr     error
	returnedStore   *multistore.Store
	called          bool
}

func (fsg *fakeStoreGetter) Get(proposalCid cid.Cid) (*multistore.Store, error) {
	fsg.lastProposalCid = proposalCid
	fsg.called = true
	return fsg.returnedStore, fsg.returnedErr
}

type fakeTransport struct{}

func (ft *fakeTransport) OpenChannel(ctx context.Context, dataSender peer.ID, channelID datatransfer.ChannelID, root ipld.Link, stor ipld.Node, msg datatransfer.Message) error {
	return nil
}

func (ft *fakeTransport) CloseChannel(ctx context.Context, chid datatransfer.ChannelID) error {
	return nil
}

func (ft *fakeTransport) SetEventHandler(events datatransfer.EventsHandler) error {
	return nil
}

func (ft *fakeTransport) CleanupChannel(chid datatransfer.ChannelID) {
}

type fakeGsTransport struct {
	datatransfer.Transport
	lastChannelID datatransfer.ChannelID
	lastLoader    ipld.Loader
	lastStorer    ipld.Storer
	called        bool
}

func (fgt *fakeGsTransport) UseStore(channelID datatransfer.ChannelID, loader ipld.Loader, storer ipld.Storer) error {
	fgt.lastChannelID = channelID
	fgt.lastLoader = loader
	fgt.lastStorer = storer
	fgt.called = true
	return nil
}
