// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	io "io"

	cid "github.com/ipfs/go-cid"
	ipld "github.com/ipld/go-ipld-prime"
	mock "github.com/stretchr/testify/mock"

	abi "github.com/filecoin-project/go-state-types/abi"
)

// PieceIO is an autogenerated mock type for the PieceIO type
type PieceIO struct {
	mock.Mock
}

// GeneratePieceCommitment provides a mock function with given fields: rt, payloadCid, selector
func (_m *PieceIO) GeneratePieceCommitment(rt abi.RegisteredSealProof, payloadCid cid.Cid, selector ipld.Node) (cid.Cid, abi.UnpaddedPieceSize, error) {
	ret := _m.Called(rt, payloadCid, selector)

	var r0 cid.Cid
	if rf, ok := ret.Get(0).(func(abi.RegisteredSealProof, cid.Cid, ipld.Node) cid.Cid); ok {
		r0 = rf(rt, payloadCid, selector)
	} else {
		r0 = ret.Get(0).(cid.Cid)
	}

	var r1 abi.UnpaddedPieceSize
	if rf, ok := ret.Get(1).(func(abi.RegisteredSealProof, cid.Cid, ipld.Node) abi.UnpaddedPieceSize); ok {
		r1 = rf(rt, payloadCid, selector)
	} else {
		r1 = ret.Get(1).(abi.UnpaddedPieceSize)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(abi.RegisteredSealProof, cid.Cid, ipld.Node) error); ok {
		r2 = rf(rt, payloadCid, selector)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// ReadPiece provides a mock function with given fields: r
func (_m *PieceIO) ReadPiece(r io.Reader) (cid.Cid, error) {
	ret := _m.Called(r)

	var r0 cid.Cid
	if rf, ok := ret.Get(0).(func(io.Reader) cid.Cid); ok {
		r0 = rf(r)
	} else {
		r0 = ret.Get(0).(cid.Cid)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(io.Reader) error); ok {
		r1 = rf(r)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
