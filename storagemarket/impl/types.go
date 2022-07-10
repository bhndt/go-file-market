package storageimpl

import (
	"bytes"
	"errors"

	"github.com/ipfs/go-cid"

	"github.com/filecoin-project/go-fil-markets/storagemarket"
)

//go:generate cbor-gen-for StorageDataTransferVoucher

var (
	// ErrWrongVoucherType means the voucher was not the correct type can validate against
	ErrWrongVoucherType = errors.New("cannot validate voucher type.")

	// ErrNoPushAccepted just means clients do not accept pushes for storage deals
	ErrNoPushAccepted = errors.New("client should not receive data for a storage deal.")

	// ErrNoPullAccepted just means providers do not accept pulls for storage deals
	ErrNoPullAccepted = errors.New("provider should not send data for a storage deal.")

	// ErrNoDeal means no active deal was found for this vouchers proposal cid
	ErrNoDeal = errors.New("no deal found for this proposal.")

	// ErrWrongPeer means that the other peer for this data transfer request does not match
	// the other peer for the deal
	ErrWrongPeer = errors.New("data Transfer peer id and Deal peer id do not match.")

	// ErrWrongPiece means that the pieceref for this data transfer request does not match
	// the one specified in the deal
	ErrWrongPiece = errors.New("base CID for deal does not match CID for piece.")

	// ErrInacceptableDealState means the deal for this transfer is not in a deal state
	// where transfer can be performed
	ErrInacceptableDealState = errors.New("deal is not a in a state where deals are accepted.")

	// DataTransferStates are the states in which it would make sense to actually start a data transfer
	DataTransferStates = []storagemarket.StorageDealStatus{storagemarket.StorageDealProposalAccepted, storagemarket.StorageDealUnknown}
)

// StorageDataTransferVoucher is the voucher type for data transfers
// used by the storage market
type StorageDataTransferVoucher struct {
	Proposal cid.Cid
}

// ToBytes converts the StorageDataTransferVoucher to raw bytes
func (dv *StorageDataTransferVoucher) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	err := dv.MarshalCBOR(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// FromBytes converts the StorageDataTransferVoucher to raw bytes
func (dv *StorageDataTransferVoucher) FromBytes(raw []byte) error {
	r := bytes.NewReader(raw)
	return dv.UnmarshalCBOR(r)
}

// Type is the unique string identifier for a StorageDataTransferVoucher
func (dv *StorageDataTransferVoucher) Type() string {
	return "StorageDataTransferVoucher"
}
