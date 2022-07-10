// Code generated by github.com/whyrusleeping/cbor-gen. DO NOT EDIT.

package piecestore

import (
	"fmt"
	"io"

	cbg "github.com/whyrusleeping/cbor-gen"
	xerrors "golang.org/x/xerrors"
)

var _ = xerrors.Errorf

func (t *PieceInfo) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write([]byte{130}); err != nil {
		return err
	}

	// t.PieceCID ([]uint8) (slice)
	if len(t.PieceCID) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.PieceCID was too long")
	}

	if _, err := w.Write(cbg.CborEncodeMajorType(cbg.MajByteString, uint64(len(t.PieceCID)))); err != nil {
		return err
	}
	if _, err := w.Write(t.PieceCID); err != nil {
		return err
	}

	// t.Deals ([]piecestore.DealInfo) (slice)
	if len(t.Deals) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.Deals was too long")
	}

	if _, err := w.Write(cbg.CborEncodeMajorType(cbg.MajArray, uint64(len(t.Deals)))); err != nil {
		return err
	}
	for _, v := range t.Deals {
		if err := v.MarshalCBOR(w); err != nil {
			return err
		}
	}
	return nil
}

func (t *PieceInfo) UnmarshalCBOR(r io.Reader) error {
	br := cbg.GetPeeker(r)

	maj, extra, err := cbg.CborReadHeader(br)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 2 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.PieceCID ([]uint8) (slice)

	maj, extra, err = cbg.CborReadHeader(br)
	if err != nil {
		return err
	}

	if extra > cbg.ByteArrayMaxLen {
		return fmt.Errorf("t.PieceCID: byte array too large (%d)", extra)
	}
	if maj != cbg.MajByteString {
		return fmt.Errorf("expected byte array")
	}
	t.PieceCID = make([]byte, extra)
	if _, err := io.ReadFull(br, t.PieceCID); err != nil {
		return err
	}
	// t.Deals ([]piecestore.DealInfo) (slice)

	maj, extra, err = cbg.CborReadHeader(br)
	if err != nil {
		return err
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("t.Deals: array too large (%d)", extra)
	}

	if maj != cbg.MajArray {
		return fmt.Errorf("expected cbor array")
	}
	if extra > 0 {
		t.Deals = make([]DealInfo, extra)
	}
	for i := 0; i < int(extra); i++ {

		var v DealInfo
		if err := v.UnmarshalCBOR(br); err != nil {
			return err
		}

		t.Deals[i] = v
	}

	return nil
}

func (t *DealInfo) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write([]byte{132}); err != nil {
		return err
	}

	// t.DealID (uint64) (uint64)
	if _, err := w.Write(cbg.CborEncodeMajorType(cbg.MajUnsignedInt, uint64(t.DealID))); err != nil {
		return err
	}

	// t.SectorID (uint64) (uint64)
	if _, err := w.Write(cbg.CborEncodeMajorType(cbg.MajUnsignedInt, uint64(t.SectorID))); err != nil {
		return err
	}

	// t.Offset (uint64) (uint64)
	if _, err := w.Write(cbg.CborEncodeMajorType(cbg.MajUnsignedInt, uint64(t.Offset))); err != nil {
		return err
	}

	// t.Length (uint64) (uint64)
	if _, err := w.Write(cbg.CborEncodeMajorType(cbg.MajUnsignedInt, uint64(t.Length))); err != nil {
		return err
	}
	return nil
}

func (t *DealInfo) UnmarshalCBOR(r io.Reader) error {
	br := cbg.GetPeeker(r)

	maj, extra, err := cbg.CborReadHeader(br)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 4 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.DealID (uint64) (uint64)

	maj, extra, err = cbg.CborReadHeader(br)
	if err != nil {
		return err
	}
	if maj != cbg.MajUnsignedInt {
		return fmt.Errorf("wrong type for uint64 field")
	}
	t.DealID = uint64(extra)
	// t.SectorID (uint64) (uint64)

	maj, extra, err = cbg.CborReadHeader(br)
	if err != nil {
		return err
	}
	if maj != cbg.MajUnsignedInt {
		return fmt.Errorf("wrong type for uint64 field")
	}
	t.SectorID = uint64(extra)
	// t.Offset (uint64) (uint64)

	maj, extra, err = cbg.CborReadHeader(br)
	if err != nil {
		return err
	}
	if maj != cbg.MajUnsignedInt {
		return fmt.Errorf("wrong type for uint64 field")
	}
	t.Offset = uint64(extra)
	// t.Length (uint64) (uint64)

	maj, extra, err = cbg.CborReadHeader(br)
	if err != nil {
		return err
	}
	if maj != cbg.MajUnsignedInt {
		return fmt.Errorf("wrong type for uint64 field")
	}
	t.Length = uint64(extra)
	return nil
}

func (t *BlockLocation) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write([]byte{130}); err != nil {
		return err
	}

	// t.RelOffset (uint64) (uint64)
	if _, err := w.Write(cbg.CborEncodeMajorType(cbg.MajUnsignedInt, uint64(t.RelOffset))); err != nil {
		return err
	}

	// t.BlockSize (uint64) (uint64)
	if _, err := w.Write(cbg.CborEncodeMajorType(cbg.MajUnsignedInt, uint64(t.BlockSize))); err != nil {
		return err
	}
	return nil
}

func (t *BlockLocation) UnmarshalCBOR(r io.Reader) error {
	br := cbg.GetPeeker(r)

	maj, extra, err := cbg.CborReadHeader(br)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 2 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.RelOffset (uint64) (uint64)

	maj, extra, err = cbg.CborReadHeader(br)
	if err != nil {
		return err
	}
	if maj != cbg.MajUnsignedInt {
		return fmt.Errorf("wrong type for uint64 field")
	}
	t.RelOffset = uint64(extra)
	// t.BlockSize (uint64) (uint64)

	maj, extra, err = cbg.CborReadHeader(br)
	if err != nil {
		return err
	}
	if maj != cbg.MajUnsignedInt {
		return fmt.Errorf("wrong type for uint64 field")
	}
	t.BlockSize = uint64(extra)
	return nil
}

func (t *PieceBlockLocation) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write([]byte{130}); err != nil {
		return err
	}

	// t.BlockLocation (piecestore.BlockLocation) (struct)
	if err := t.BlockLocation.MarshalCBOR(w); err != nil {
		return err
	}

	// t.PieceCID ([]uint8) (slice)
	if len(t.PieceCID) > cbg.ByteArrayMaxLen {
		return xerrors.Errorf("Byte array in field t.PieceCID was too long")
	}

	if _, err := w.Write(cbg.CborEncodeMajorType(cbg.MajByteString, uint64(len(t.PieceCID)))); err != nil {
		return err
	}
	if _, err := w.Write(t.PieceCID); err != nil {
		return err
	}
	return nil
}

func (t *PieceBlockLocation) UnmarshalCBOR(r io.Reader) error {
	br := cbg.GetPeeker(r)

	maj, extra, err := cbg.CborReadHeader(br)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 2 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.BlockLocation (piecestore.BlockLocation) (struct)

	{

		if err := t.BlockLocation.UnmarshalCBOR(br); err != nil {
			return err
		}

	}
	// t.PieceCID ([]uint8) (slice)

	maj, extra, err = cbg.CborReadHeader(br)
	if err != nil {
		return err
	}

	if extra > cbg.ByteArrayMaxLen {
		return fmt.Errorf("t.PieceCID: byte array too large (%d)", extra)
	}
	if maj != cbg.MajByteString {
		return fmt.Errorf("expected byte array")
	}
	t.PieceCID = make([]byte, extra)
	if _, err := io.ReadFull(br, t.PieceCID); err != nil {
		return err
	}
	return nil
}

func (t *CIDInfo) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write([]byte{130}); err != nil {
		return err
	}

	// t.CID (cid.Cid) (struct)

	if err := cbg.WriteCid(w, t.CID); err != nil {
		return xerrors.Errorf("failed to write cid field t.CID: %w", err)
	}

	// t.PieceBlockLocations ([]piecestore.PieceBlockLocation) (slice)
	if len(t.PieceBlockLocations) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.PieceBlockLocations was too long")
	}

	if _, err := w.Write(cbg.CborEncodeMajorType(cbg.MajArray, uint64(len(t.PieceBlockLocations)))); err != nil {
		return err
	}
	for _, v := range t.PieceBlockLocations {
		if err := v.MarshalCBOR(w); err != nil {
			return err
		}
	}
	return nil
}

func (t *CIDInfo) UnmarshalCBOR(r io.Reader) error {
	br := cbg.GetPeeker(r)

	maj, extra, err := cbg.CborReadHeader(br)
	if err != nil {
		return err
	}
	if maj != cbg.MajArray {
		return fmt.Errorf("cbor input should be of type array")
	}

	if extra != 2 {
		return fmt.Errorf("cbor input had wrong number of fields")
	}

	// t.CID (cid.Cid) (struct)

	{

		c, err := cbg.ReadCid(br)
		if err != nil {
			return xerrors.Errorf("failed to read cid field t.CID: %w", err)
		}

		t.CID = c

	}
	// t.PieceBlockLocations ([]piecestore.PieceBlockLocation) (slice)

	maj, extra, err = cbg.CborReadHeader(br)
	if err != nil {
		return err
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("t.PieceBlockLocations: array too large (%d)", extra)
	}

	if maj != cbg.MajArray {
		return fmt.Errorf("expected cbor array")
	}
	if extra > 0 {
		t.PieceBlockLocations = make([]PieceBlockLocation, extra)
	}
	for i := 0; i < int(extra); i++ {

		var v PieceBlockLocation
		if err := v.UnmarshalCBOR(br); err != nil {
			return err
		}

		t.PieceBlockLocations[i] = v
	}

	return nil
}
