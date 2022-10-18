package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/omnibtc/go-hippo-sdk/aggregator"
	"github.com/omnibtc/go-hippo-sdk/aggregator/auxamm"
	"github.com/omnibtc/go-hippo-sdk/aggregator/base"
	"github.com/omnibtc/go-hippo-sdk/aggregator/basiq"
	"github.com/omnibtc/go-hippo-sdk/aggregator/coinlist"
	"github.com/omnibtc/go-hippo-sdk/aggregator/pontem"
	"github.com/omnibtc/go-hippo-sdk/contract"
	"github.com/omnibtc/go-hippo-sdk/types"
)

const TestNode = "https://fullnode.testnet.aptoslabs.com"
const basiqPoolAddress = "0x4885b08864b81ca42b19c38fff2eb958b5e312b1ec366014d4afff2775c19aab"
const auxPoolAddress = "0xe1d61154f57bbbb256bb6e3ea786102e7d5c9af481cb4d11b11e579f98218f27"
const pontemAddress = "0x385068db10693e06512ed54b1e6e8f1fb9945bb7a78c28a45585939ce953f99e"

func main() {
	client, err := aptosclient.Dial(context.Background(), TestNode)
	panicErr(err)

	coinListApp := contract.NewDevCoinListApp()
	coinListClient, err := coinlist.LoadCoinListClient(contract.App{
		CoinList: coinListApp,
	})
	panicErr(err)

	aggr := aggregator.NewTradeAggregator(
		contract.App{
			CoinList: coinListApp,
		},
		types.SimulationKeys{},
		[]base.TradingPoolProvider{
			basiq.NewPoolProvider(client, basiqPoolAddress, coinListClient),
			auxamm.NewPoolProvider(client, auxPoolAddress, coinListClient),
			pontem.NewPoolProvider(client, pontemAddress, coinListClient),
		},
	)
	coinX, ok := coinListClient.GetCoinInfoByFullName("0x498d8926f16eb9ca90cab1b3a26aa6f97a080b3fcbe6e83ae150b7243a00fb68::devnet_coins::DevnetBTC")
	if !ok {
		panic("coinx not found")
	}
	coinY, ok := coinListClient.GetCoinInfoByFullName("0x498d8926f16eb9ca90cab1b3a26aa6f97a080b3fcbe6e83ae150b7243a00fb68::devnet_coins::DevnetUSDT")
	if !ok {
		panic("coiny not found")
	}
	inputAmount := big.NewInt(10000000000)
	quotes, err := aggr.GetQuotes(inputAmount, coinX, coinY, 3, false, false)
	panicErr(err)

	fmt.Printf("quote size: %d\n", len(quotes))

	for i, q := range quotes {
		fmt.Printf("quote:%d\n", i)
		fmt.Printf("Path: ")
		for _, p := range q.Route.Steps {
			fmt.Printf(" %s ", p.Pool.DexType().Name())
		}
		fmt.Printf("out: %s\n", ((*big.Int)(q.Quote.OutputAmount)).String())
		fmt.Printf("%v\n", q.Route.MakePayload(inputAmount, quotes[0].Quote.OutputAmount))
	}
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
