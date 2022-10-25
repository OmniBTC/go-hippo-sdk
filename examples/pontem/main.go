package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/omnibtc/go-hippo-sdk/aggregator"
	"github.com/omnibtc/go-hippo-sdk/aggregator/base"
	"github.com/omnibtc/go-hippo-sdk/aggregator/coinlist"
	"github.com/omnibtc/go-hippo-sdk/aggregator/pontem"
	"github.com/omnibtc/go-hippo-sdk/contract"
	"github.com/omnibtc/go-hippo-sdk/types"
)

const TestNode = "https://fullnode.mainnet.aptoslabs.com"
const pontemAddress = "0x05a97986a9d031c4567e15b797be516910cfcb4156312482efc6a19c0a30c948"

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
		[]base.TradingPoolProvider{pontem.NewPoolProvider(client, pontemAddress, coinListClient)},
	)
	coinX, ok := coinListClient.GetCoinInfoByFullName("0xf22bede237a07e121b56d91a491eb7bcdfd1f5907926a9e58338f964a01b17fa::asset::USDC")
	if !ok {
		panic("coiny not found")
	}
	coinY, ok := coinListClient.GetCoinInfoByFullName("0xf22bede237a07e121b56d91a491eb7bcdfd1f5907926a9e58338f964a01b17fa::asset::USDT")
	if !ok {
		panic("coinx not found")
	}
	inputAmount := big.NewInt(10000000)
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
