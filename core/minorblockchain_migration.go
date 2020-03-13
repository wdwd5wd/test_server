// Implement account migration
package core

import (
	"github.com/QuarkChain/goquarkchain/core/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

// Retrieve & remove all queued & pending transcations originated from given account.
// Returned transcations are sorted by nonce.
func (m *MinorBlockChain) PopAccountTranscations(account common.Address) types.Transactions {
	m.txPool.mu.Lock()
	defer m.txPool.mu.Unlock()

	txs := newTxSortedMap()

	pending := m.txPool.pending[account]
	if pending != nil {
		log.Debug("found pending txs", "size", pending.Len())
		for _, tx := range pending.txs.items {
			txs.Put(tx)
		}
		delete(m.txPool.pending, account)
	}
	queued := m.txPool.queue[account]
	if queued != nil {
		log.Debug("found queued txs", "size", queued.Len())
		for _, tx := range queued.txs.items {
			txs.Put(tx)
		}
		delete(m.txPool.queue, account)
	}

	return txs.Flatten()
}
