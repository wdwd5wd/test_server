// Implement account migration
package shard

import (
	"github.com/QuarkChain/goquarkchain/core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// Retrieve & remove all queued & pending transcations originated from given account on
// that shard. Mining will restart. Returned transcations are sorted by nonce.
func (s *ShardBackend) PopAccountTranscations(account common.Address) types.Transactions {
	txs := s.MinorBlockChain.PopAccountTranscations(account)
	log.Debug("Transcations poped from pool", "size", len(txs))

	// Restart mining
	if s.miner.IsMining() {
		log.Debug("Restart mining")
		s.miner.SetMining(false)
		s.miner.SetMining(true)
	}
	return txs
}
