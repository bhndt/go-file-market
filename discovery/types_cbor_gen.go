// Code generated by github.com/whyrusleeping/cbor-gen. DO NOT EDIT.

package discovery

import (
	"fmt"
	"io"
	"sort"

	retrievalmarket "github.com/filecoin-project/go-fil-markets/retrievalmarket"
	cid "github.com/ipfs/go-cid"
	cbg "github.com/whyrusleeping/cbor-gen"
	xerrors "golang.org/x/xerrors"
)

var _ = xerrors.Errorf
var _ = cid.Undef
var _ = sort.Sort

func (t *RetrievalPeers) MarshalCBOR(w io.Writer) error {
	if t == nil {
		_, err := w.Write(cbg.CborNull)
		return err
	}
	if _, err := w.Write([]byte{161}); err != nil {
		return err
	}

	scratch := make([]byte, 9)

	// t.Peers ([]retrievalmarket.RetrievalPeer) (slice)
	if len("Peers") > cbg.MaxLength {
		return xerrors.Errorf("Value in field \"Peers\" was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajTextString, uint64(len("Peers"))); err != nil {
		return err
	}
	if _, err := io.WriteString(w, string("Peers")); err != nil {
		return err
	}

	if len(t.Peers) > cbg.MaxLength {
		return xerrors.Errorf("Slice value in field t.Peers was too long")
	}

	if err := cbg.WriteMajorTypeHeaderBuf(scratch, w, cbg.MajArray, uint64(len(t.Peers))); err != nil {
		return err
	}
	for _, v := range t.Peers {
		if err := v.MarshalCBOR(w); err != nil {
			return err
		}
	}
	return nil
}

func (t *RetrievalPeers) UnmarshalCBOR(r io.Reader) error {
	*t = RetrievalPeers{}

	br := cbg.GetPeeker(r)
	scratch := make([]byte, 8)

	maj, extra, err := cbg.CborReadHeaderBuf(br, scratch)
	if err != nil {
		return err
	}
	if maj != cbg.MajMap {
		return fmt.Errorf("cbor input should be of type map")
	}

	if extra > cbg.MaxLength {
		return fmt.Errorf("RetrievalPeers: map struct too large (%d)", extra)
	}

	var name string
	n := extra

	for i := uint64(0); i < n; i++ {

		{
			sval, err := cbg.ReadStringBuf(br, scratch)
			if err != nil {
				return err
			}

			name = string(sval)
		}

		switch name {
		// t.Peers ([]retrievalmarket.RetrievalPeer) (slice)
		case "Peers":

			maj, extra, err = cbg.CborReadHeaderBuf(br, scratch)
			if err != nil {
				return err
			}

			if extra > cbg.MaxLength {
				return fmt.Errorf("t.Peers: array too large (%d)", extra)
			}

			if maj != cbg.MajArray {
				return fmt.Errorf("expected cbor array")
			}

			if extra > 0 {
				t.Peers = make([]retrievalmarket.RetrievalPeer, extra)
			}

			for i := 0; i < int(extra); i++ {

				var v retrievalmarket.RetrievalPeer
				if err := v.UnmarshalCBOR(br); err != nil {
					return err
				}

				t.Peers[i] = v
			}

		default:
			// Field doesn't exist on this type, so ignore it
			cbg.ScanForLinks(r, func(cid.Cid) {})
		}
	}

	return nil
}
