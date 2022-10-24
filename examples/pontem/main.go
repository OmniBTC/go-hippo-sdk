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
const pontemAddress = "0x5a97986a9d031c4567e15b797be516910cfcb4156312482efc6a19c0a30c948"

func main() {
	client, err := aptosclient.Dial(context.Background(), TestNode)
	panicErr(err)

	coinListApp := contract.NewCustomCoinListApp([]types.CoinInfo{
		{
			// 0x1::aptos_coin::AptosCoin
			Name:     "APT",
			Symbol:   "APT",
			Decimals: 8,
			TokenType: &types.StructTag{
				Address: "0x1",
				Module:  "aptos_coin",
				Name:    "AptosCoin",
			},
		},
		{
			// 0xa2eda21a58856fda86451436513b867c97eecb4ba099da5775520e0f7492e852::coin::T
			Name:     "USDC",
			Symbol:   "USDC",
			Decimals: 6,
			TokenType: &types.StructTag{
				Address: "0xc7160b1c2415d19a88add188ec726e62aab0045f0aed798106a2ef2994a9101e",
				Module:  "coin",
				Name:    "T",
			},
		},
		{
			// 0xa2eda21a58856fda86451436513b867c97eecb4ba099da5775520e0f7492e852::coin::T
			Name:     "USDT",
			Symbol:   "USDT",
			Decimals: 6,
			TokenType: &types.StructTag{
				Address: "0xa2eda21a58856fda86451436513b867c97eecb4ba099da5775520e0f7492e852",
				Module:  "coin",
				Name:    "T",
			},
		},
	})
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
	coinY, ok := coinListClient.GetCoinInfoByFullName("0xc7160b1c2415d19a88add188ec726e62aab0045f0aed798106a2ef2994a9101e::coin::T")
	if !ok {
		panic("coinx not found")
	}
	coinX, ok := coinListClient.GetCoinInfoByFullName("0x1::aptos_coin::AptosCoin")
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
