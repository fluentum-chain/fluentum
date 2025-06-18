package psql

import (
	"github.com/fluentum-chain/fluentum/state/indexer"
	"github.com/fluentum-chain/fluentum/state/txindex"
)

var (
	_ indexer.BlockIndexer = BackportBlockIndexer{}
	_ txindex.TxIndexer    = BackportTxIndexer{}
)
