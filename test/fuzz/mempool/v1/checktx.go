package v1

import (
	"github.com/fluentum-chain/fluentum/abci/example/kvstore"
	"github.com/fluentum-chain/fluentum/config"
	"github.com/fluentum-chain/fluentum/libs/log"
	mempl "github.com/fluentum-chain/fluentum/mempool"
	"github.com/fluentum-chain/fluentum/proxy"

	mempoolv1 "github.com/fluentum-chain/fluentum/mempool/v1"
)

var mempool mempl.Mempool

func init() {
	app := kvstore.NewApplication()
	cc := proxy.NewLocalClientCreator(app)
	appConnMem, _ := cc.NewABCIClient()
	err := appConnMem.Start()
	if err != nil {
		panic(err)
	}
	cfg := config.DefaultMempoolConfig()
	cfg.Broadcast = false
	log := log.NewNopLogger()
	appConnMempool := proxy.NewAppConnMempool(appConnMem)
	mempool = mempoolv1.NewTxMempool(log, cfg, appConnMempool, 0)
}

func Fuzz(data []byte) int {

	err := mempool.CheckTx(data, nil, mempl.TxInfo{})
	if err != nil {
		return 0
	}

	return 1
}
