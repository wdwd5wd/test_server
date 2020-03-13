// Implement account migration
package shard

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

func (s *ShardBackend) MigrateAccountToOtherShard(account common.Address, newFullShardKey uint32) error {
	log.Debug("Migrate", "account", account)
	txs := s.MinorBlockChain.PopAccountTranscations(account)
	log.Debug("Transcations poped from pool", "size", len(txs))

	// Restart mining
	if s.miner.IsMining() {
		log.Debug("Restart mining")
		s.miner.SetMining(false)
		s.miner.SetMining(true)
	}

	// Change FromFullShardKey to the newFullShardKey for every tx
	for _, tx := range txs {
		log.Debug("Migrate to new shard key", "tx", tx)
		// TODO: modify shard id of tx
	}

	// TODO: send transcations to master?

	return nil
}
