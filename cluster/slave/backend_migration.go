package slave

import (
	"errors"
	"fmt"

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

	fromShard.RestartMining(false)

	txs := fromShard.PopAccountTranscations(account)

	// 由于grpc的传输大小限制，交易数量过多则会被拆分传输
	var txsChunk types.Transactions
	MaxTxLen := 20000

	// Set new FromFullShardKey
	newFromFullShardKey := types.Uint32(toShardKey)
	for index, tx := range txs {
		tx.EvmTx.TxData.FromFullShardKey = &newFromFullShardKey

		txsChunk = append(txsChunk, tx)

		if (index+1) >= MaxTxLen && (index+1)%MaxTxLen == 0 {
			fmt.Println("txsChunk size:", len(txsChunk))
			// Boradcast txs
			err := s.connManager.BroadcastTransactions("!MIGRATION!", toShardKey, txsChunk)
			if err != nil {
				log.Warn("fail to broadcast migrated txs")
				fmt.Println(err)
			}
			txsChunk = txsChunk[:0]
		}
	}

	// Boradcast txs
	err := s.connManager.BroadcastTransactions("!MIGRATION!", toShardKey, txsChunk)
	if err != nil {
		log.Warn("fail to broadcast migrated txs")
		fmt.Println(err)
	}
	return nil
}

func (s *SlaveBackend) MigrationEnded(end bool, fromShardKey uint32) {

	if end {
		fromShard := s.GetShard(fromShardKey)
		if fromShard == nil {
			fmt.Println("the shard does not exist")
		}
		fromShard.RestartMining(end)
	}

}
