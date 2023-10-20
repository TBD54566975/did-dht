package dht

import (
	"context"
	"crypto/sha1"
	"errors"
	"math"

	"github.com/anacrolix/log"
	"github.com/anacrolix/torrent/bencode"

	"github.com/anacrolix/dht/v2"
	"github.com/anacrolix/dht/v2/bep44"
	"github.com/anacrolix/dht/v2/krpc"
	"github.com/anacrolix/dht/v2/traversal"
)

// Copied from https://github.com/anacrolix/dht/blob/master/exts/getput/getput.go and modified
// to return signature data

type FullGetResult struct {
	Seq     int64
	V       bencode.Bytes
	Sig     [64]byte
	Mutable bool
}

func startGetTraversal(
	target bep44.Target, s *dht.Server, seq *int64, salt []byte,
) (
	vChan chan FullGetResult, op *traversal.Operation, err error,
) {
	vChan = make(chan FullGetResult)
	op = traversal.Start(traversal.OperationInput{
		Alpha:  15,
		Target: target,
		DoQuery: func(ctx context.Context, addr krpc.NodeAddr) traversal.QueryResult {
			logger := log.ContextLogger(ctx)
			res := s.Get(ctx, dht.NewAddr(addr.UDP()), target, seq, dht.QueryRateLimiting{})
			err := res.ToError()
			if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, dht.TransactionTimeout) {
				logger.Levelf(log.Debug, "error querying %v: %v", addr, err)
			}
			if r := res.Reply.R; r != nil {
				rv := r.V
				bv := rv
				if sha1.Sum(bv) == target {
					select {
					case vChan <- FullGetResult{
						V:   rv,
						Sig: r.Sig,
					}:
					case <-ctx.Done():
					}
				} else if sha1.Sum(append(r.K[:], salt...)) == target && bep44.Verify(r.K[:], salt, *r.Seq, bv, r.Sig[:]) {
					select {
					case vChan <- FullGetResult{
						Seq: *r.Seq,
						V:   rv,
						Sig: r.Sig,
					}:
					case <-ctx.Done():
					}
				} else if rv != nil {
					logger.Levelf(log.Debug, "get response item hash didn't match target: %q", rv)
				}
			}
			tqr := res.TraversalQueryResult(addr)
			// Filter replies from nodes that don't have a string token. This doesn't look prettier
			// with generics. "The token value should be a short binary string." ¯\_(ツ)_/¯ (BEP 5).
			tqr.ClosestData, _ = tqr.ClosestData.(string)
			if tqr.ClosestData == nil {
				tqr.ResponseFrom = nil
			}
			return tqr
		},
		NodeFilter: s.TraversalNodeFilter,
	})
	nodes, err := s.TraversalStartingNodes()
	op.AddNodes(nodes)
	return
}

func Get(
	ctx context.Context, target bep44.Target, s *dht.Server, seq *int64, salt []byte,
) (
	ret FullGetResult, stats *traversal.Stats, err error,
) {
	vChan, op, err := startGetTraversal(target, s, seq, salt)
	if err != nil {
		return
	}
	ret.Seq = math.MinInt64
	gotValue := false
receiveResults:
	select {
	case <-op.Stalled():
		if !gotValue {
			err = errors.New("value not found")
		}
	case v := <-vChan:
		log.ContextLogger(ctx).Levelf(log.Debug, "received %#v", v)
		gotValue = true
		if !v.Mutable {
			ret = v
			break
		}
		if v.Seq >= ret.Seq {
			ret = v
		}
		goto receiveResults
	case <-ctx.Done():
		err = ctx.Err()
	}
	op.Stop()
	stats = op.Stats()
	return
}
