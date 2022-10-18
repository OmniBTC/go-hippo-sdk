package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/omnibtc/go-hippo-sdk/aggregator"
	"github.com/omnibtc/go-hippo-sdk/aggregator/base"
	"github.com/omnibtc/go-hippo-sdk/aggregator/basiq"
	"github.com/omnibtc/go-hippo-sdk/aggregator/coinlist"
	"github.com/omnibtc/go-hippo-sdk/contract"
	"github.com/omnibtc/go-hippo-sdk/types"
)

const poolAddress = "0x4885b08864b81ca42b19c38fff2eb958b5e312b1ec366014d4afff2775c19aab"
const TestNode = "https://fullnode.testnet.aptoslabs.com"

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
		[]base.TradingPoolProvider{basiq.NewPoolProvider(client, poolAddress, coinListClient)},
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

	for _, q := range quotes {
		fmt.Printf("out: %s\n", ((*big.Int)(q.Quote.OutputAmount)).String())
	}

	if len(quotes) > 0 {
		fmt.Printf("%v\n", quotes[0].Route.MakePayload(inputAmount, quotes[0].Quote.OutputAmount))
	}
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
