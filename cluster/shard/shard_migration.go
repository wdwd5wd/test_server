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
	// Restart mining
	// s.minerSem.Release(1)
	//s.mu.Lock()

	// if s.miner.IsMining() {
	// 	log.Debug("Terminate mining")
	// 	s.miner.SetMining(false)
	// }
	//s.mu.Unlock()

	txs := s.MinorBlockChain.PopAccountTranscations(account)
	log.Debug("Transcations poped from pool", "size", len(txs))

	//s.mu.Lock()
	// if !s.miner.IsMining() {
	// 	log.Debug("Restart mining")
	// 	s.miner.SetMining(true)
	// }
	//s.mu.Unlock()

	return txs
}

func (s *ShardBackend) RestartMining(ok bool) {

	s.mu.Lock()

	if s.miner.IsMining() && !ok {
		log.Debug("Terminate mining")
		s.miner.SetMining(false)
	}

	if !s.miner.IsMining() && ok {
		log.Debug("Restart mining")
		s.miner.SetMining(true)
	}

	s.mu.Unlock()
}

// Retrieve all queued & pending transcations originated from given account on
// that shard. Mining will restart. Returned transcations are sorted by nonce.
func (s *ShardBackend) CountAccountTranscations(account common.Address) types.Transactions {
	// Restart mining
	// if s.miner.IsMining() {
	// 	log.Debug("Terminate mining")
	// 	s.miner.SetMining(false)
	// }

	txs := s.MinorBlockChain.CountBroadcastTxs(account)
	log.Debug("Transcations COUNTED from pool", "size", len(txs))

	// if !s.miner.IsMining() {
	// 	log.Debug("Restart mining")
	// 	s.miner.SetMining(true)
	// }

	return txs
}
