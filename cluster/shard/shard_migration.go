// Implement account migration
package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

func (s *ShardBackend) MigrateAccountToOtherShard(account common.Address) error {
	log.Debug("Migrate", "account", account)
	txs := s.MinorBlockChain.txPool.popAccountTranscations(account)
	log.Debug("Transcations poped from pool", "size", len(txs))

	if s.miner.IsMining() {
		log.Debug("Restart mining")
		s.miner.SetMining(false)
		s.miner.SetMining(true)
	}

	// TODO: send transcations to master?

	return nil
}
