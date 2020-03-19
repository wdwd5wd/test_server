package slave

import (
	"errors"

	"github.com/QuarkChain/goquarkchain/core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// Remove all pending & queued transcations of the account, and send them to
// the other shard.
func (s *SlaveBackend) MigrateAccountToOtherShard(account common.Address,
	fromShardKey uint32, toShardKey uint32) error {
	log.Debug("Migrate", "account", account, "from", fromShardKey, "to", toShardKey)

	fromShard := s.GetShard(fromShardKey)
	if fromShard == nil {
		return errors.New("the shard does not exist")
	}
	txs := fromShard.PopAccountTranscations(account)

	// Set new FromFullShardKey
	newFromFullShardKey := types.Uint32(toShardKey)
	for _, tx := range txs {
		tx.EvmTx.TxData.FromFullShardKey = &newFromFullShardKey
	}

	// Boradcast txs
	err := s.connManager.BroadcastTransactions("", toShardKey, txs)
	if err != nil {
		log.Warn("fail to broadcast migrated txs")
		return err
	}
	return nil
}
