package main

import (
	"context"
	"fmt"
	"math/big"

	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/omnibtc/go-hippo-sdk/aggregator"
	"github.com/omnibtc/go-hippo-sdk/aggregator/auxamm"
	"github.com/omnibtc/go-hippo-sdk/aggregator/base"
	"github.com/omnibtc/go-hippo-sdk/aggregator/coinlist"
	"github.com/omnibtc/go-hippo-sdk/contract"
	"github.com/omnibtc/go-hippo-sdk/types"
)

const poolAddress = "0xbd35135844473187163ca197ca93b2ab014370587bb0ed3befff9e902d6bb541"
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
		[]base.TradingPoolProvider{auxamm.NewPoolProvider(client, poolAddress, coinListClient)},
	)
	coinX, ok := coinListClient.GetCoinInfoByFullName("0x1::aptos_coin::AptosCoin")
	if !ok {
		panic("coinx not found")
	}
	coinY, ok := coinListClient.GetCoinInfoByFullName("0x881ac202b1f1e6ad4efcff7a1d0579411533f2502417a19211cfc49751ddb5f4::coin::MOJO")
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
