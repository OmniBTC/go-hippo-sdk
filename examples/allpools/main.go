package main

import (
	"context"
	"fmt"
	"github.com/omnibtc/go-hippo-sdk/aggregator/anime"
	"github.com/omnibtc/go-hippo-sdk/aggregator/aptosswap"
	"github.com/omnibtc/go-hippo-sdk/aggregator/pancake"
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

const TestNode = "https://fullnode.mainnet.aptoslabs.com"
const basiqPoolAddress = "0x4885b08864b81ca42b19c38fff2eb958b5e312b1ec366014d4afff2775c19aab"
const auxPoolAddress = "0xbd35135844473187163ca197ca93b2ab014370587bb0ed3befff9e902d6bb541"
const pontemAddress = "0x05a97986a9d031c4567e15b797be516910cfcb4156312482efc6a19c0a30c948"
const aptosPoolAddress = "0xa5d3ac4d429052674ed38adc62d010e52d7c24ca159194d17ddc196ddb7e480b"
const animePoolAddress = "0x796900ebe1a1a54ff9e932f19c548f5c1af5c6e7d34965857ac2f7b1d1ab2cbf"
const pancakePoolAddress = "0xc7efb4076dbe143cbcd98cfaaa929ecfc8f299203dfff63b95ccb6bfe19850fa"

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
			aptosswap.NewPoolProvider(client, aptosPoolAddress, coinListClient),
			anime.NewPoolProvider(client, animePoolAddress, coinListClient),
			pancake.NewPoolProvider(client, pancakePoolAddress, coinListClient),
		},
	)
	coinX, ok := coinListClient.GetCoinInfoByFullName("0x1::aptos_coin::AptosCoin")
	if !ok {
		panic("coinx not found")
	}
	coinY, ok := coinListClient.GetCoinInfoByFullName("0xf22bede237a07e121b56d91a491eb7bcdfd1f5907926a9e58338f964a01b17fa::asset::USDC")
	if !ok {
		panic("coiny not found")
	}
	inputAmount := big.NewInt(100000000)
	quotes, err := aggr.GetQuotes(inputAmount, coinX, coinY, 3, false, false)
	panicErr(err)

	fmt.Printf("quote size: %d\n", len(quotes))

	for i, q := range quotes {
		if i == 10 {
			return
		}
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
