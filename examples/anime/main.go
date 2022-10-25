package main

import (
	"context"
	"fmt"
	"github.com/omnibtc/go-hippo-sdk/aggregator/anime"
	"math/big"

	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/omnibtc/go-hippo-sdk/aggregator"
	"github.com/omnibtc/go-hippo-sdk/aggregator/base"
	"github.com/omnibtc/go-hippo-sdk/aggregator/coinlist"
	"github.com/omnibtc/go-hippo-sdk/contract"
	"github.com/omnibtc/go-hippo-sdk/types"
)

const poolAddress = "0x796900ebe1a1a54ff9e932f19c548f5c1af5c6e7d34965857ac2f7b1d1ab2cbf"
const TestNode = "https://fullnode.mainnet.aptoslabs.com"

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
		[]base.TradingPoolProvider{anime.NewPoolProvider(client, poolAddress, coinListClient)},
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
