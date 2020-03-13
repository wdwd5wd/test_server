// Implement account migration
package core

import (
	"github.com/QuarkChain/goquarkchain/core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// Retrieve & remove all queued & pending transcations originated from given account.
// Returned transcations are sorted by nonce.
func (pool *TxPool) popAccountTranscations(account common.Address) types.Transactions {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	txs := newTxSortedMap()

	pending := pool.pending[account]
	if pending != nil {
		log.Debug("found pending txs", "size", pending.Len())
		for _, tx := range pending.txs.items {
			txs.Put(tx)
		}
		delete(pool.pending, account)
	}
	queued := pool.queue[account]
	if queued != nil {
		log.Debug("found queued txs", "size", queued.Len())
		for _, tx := range queued.txs.items {
			txs.Put(tx)
		}
		delete(pool.queue, account)
	}

	return txs.Flatten()
}
