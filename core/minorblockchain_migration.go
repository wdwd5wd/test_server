// Implement account migration
package core

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

func (m *MinorBlockChain) MigrateAccountToOtherShard(account common.Address) error {
	log.Debug("Migrate", "account", account)
	txs := m.txPool.popAccountTranscations(account)
	log.Debug("Transcations poped from pool", "size", len(txs))

	// TODO: send transcations to master?

	return nil
}
