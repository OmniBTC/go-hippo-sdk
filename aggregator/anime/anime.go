package anime

import (
	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/coming-chat/go-aptos/aptostypes"
	"github.com/omnibtc/go-hippo-sdk/aggregator/base"
	"github.com/omnibtc/go-hippo-sdk/aggregator/coinlist"
	"github.com/omnibtc/go-hippo-sdk/types"
	"math/big"
	"strings"
)

type LiquidityPool struct {
	CoinXReserve         types.Coin
	CoinYReserve         types.Coin
	LastBlockTimestamp   int64
	LastPriceXCumulative *big.Int
	LastPriceYCumulative *big.Int
	KLast                *big.Int
	Locked               bool
}

func NewLiquidityPool(resource aptostypes.AccountResource) *LiquidityPool {
	data := resource.Data
	coinXValue, b := big.NewInt(0).SetString(data["coin_x_reserve"].(map[string]interface{})["value"].(string), 10)
	if !b {
		return nil
	}
	coinXReserve := types.Coin{
		Value: coinXValue,
	}

	coinYValue, b := big.NewInt(0).SetString(data["coin_y_reserve"].(map[string]interface{})["value"].(string), 10)
	if !b {
		return nil
	}
	coinYReserve := types.Coin{
		Value: coinYValue,
	}
	if coinXValue.Cmp(big.NewInt(0)) == 0 || coinYValue.Cmp(big.NewInt(0)) == 0 {
		return nil
	}

	kLast, b := big.NewInt(0).SetString(data["k_last"].(string), 10)
	if !b {
		return nil
	}
	lastBlockTimestamp, b := big.NewInt(0).SetString(data["last_block_timestamp"].(string), 10)
	if !b {
		return nil
	}

	lastPriceXCumulative, b := big.NewInt(0).SetString(data["last_price_x_cumulative"].(string), 10)
	if !b {
		return nil
	}
	lastPriceYCumulative, b := big.NewInt(0).SetString(data["last_price_y_cumulative"].(string), 10)
	if !b {
		return nil
	}
	locked := data["locked"].(bool)

	return &LiquidityPool{
		CoinXReserve:         coinXReserve,
		CoinYReserve:         coinYReserve,
		LastBlockTimestamp:   lastBlockTimestamp.Int64(),
		LastPriceXCumulative: lastPriceXCumulative,
		LastPriceYCumulative: lastPriceYCumulative,
		Locked:               locked,
		KLast:                kLast,
	}
}

type AnimeTradingPool struct {
	OwnerAddr  string
	_xCoinInfo types.CoinInfo
	_yCoinInfo types.CoinInfo
	Tag        types.StructTag
	Pool       LiquidityPool
}

func NewAnimeTradingPool(owner string, xCoin, yCoin types.CoinInfo, tag types.StructTag, resource aptostypes.AccountResource) *AnimeTradingPool {
	pool := NewLiquidityPool(resource)
	if pool == nil {
		return nil
	}
	return &AnimeTradingPool{
		OwnerAddr:  owner,
		_xCoinInfo: xCoin,
		_yCoinInfo: yCoin,
		Tag:        tag,
		Pool:       *pool,
	}
}

func (a *AnimeTradingPool) DexType() base.DexType {
	return base.AnimeSwap
}

func (a *AnimeTradingPool) PoolType() base.PoolType {
	return 0
}

func (a *AnimeTradingPool) IsRoutable() bool {
	return true
}

func (a *AnimeTradingPool) XCoinInfo() types.CoinInfo {
	return a._xCoinInfo
}

func (a *AnimeTradingPool) YCoinInfo() types.CoinInfo {
	return a._yCoinInfo
}

func (a *AnimeTradingPool) IsStateLoaded() bool {
	return true
}

func (a *AnimeTradingPool) GetPrice() base.PriceType {
	panic("not implemented")
}

func (a *AnimeTradingPool) GetQuote(inputAmount base.TokenAmount, isXToY bool) base.QuoteType {
	if !a.IsStateLoaded() {
		panic("anime pool not loaded")
	}
	inputTokenInfo := a._xCoinInfo
	outputTokenInfo := a._yCoinInfo
	inputReserve := a.Pool.CoinXReserve.Value
	outputReserve := a.Pool.CoinYReserve.Value
	if !isXToY {
		inputTokenInfo, outputTokenInfo = outputTokenInfo, inputTokenInfo
		inputReserve, outputReserve = outputReserve, inputReserve
	}
	coinOutAmt := getAmountOut(inputAmount, inputReserve, outputReserve, big.NewInt(30))
	return base.QuoteType{
		InputSymbol:  inputTokenInfo.Symbol,
		OutputSymbol: outputTokenInfo.Symbol,
		InputAmount:  inputAmount,
		OutputAmount: coinOutAmt,
	}
}

func (a *AnimeTradingPool) GetTagE() types.TokenType {
	return types.U8
}

func (a *AnimeTradingPool) MakePayload(base.TokenAmount, base.TokenAmount) types.EntryFunctionPayload {
	panic("not implemented")
}

func getAmountOut(amountIn, reserveIn, reserveOut, swapFee *big.Int) *big.Int {
	var temp bool
	if amountIn.Cmp(big.NewInt(0)) < 0 {
		panic("insufficient input amount")
	}
	if reserveIn.Cmp(big.NewInt(0)) > 0 {
		temp = reserveOut.Cmp(big.NewInt(0)) > 0
	}
	if !temp {
		panic("insufficient liquidity")
	}
	amountInWithFee := new(big.Int).Mul(amountIn, new(big.Int).Sub(big.NewInt(10000), swapFee))
	numerator := new(big.Int).Mul(amountInWithFee, reserveOut)
	denominator := new(big.Int).Add(new(big.Int).Mul(reserveIn, big.NewInt(10000)), amountInWithFee)
	amountOut := new(big.Int).Div(numerator, denominator)
	return amountOut
}

type AnimePoolProvider struct {
	client         *aptosclient.RestClient
	ownerAddress   string
	coinListClient *coinlist.CoinListClient
}

func NewPoolProvider(client *aptosclient.RestClient, ownerAddress string, coinListClient *coinlist.CoinListClient) base.TradingPoolProvider {
	return &AnimePoolProvider{
		client:         client,
		ownerAddress:   ownerAddress,
		coinListClient: coinListClient,
	}
}

func (p *AnimePoolProvider) LoadPoolList() []base.TradingPool {
	poolList := make([]base.TradingPool, 0)
	resources, err := p.client.GetAccountResources(p.ownerAddress)
	if err != nil {
		return poolList
	}

	for _, resource := range resources {
		if !strings.Contains(resource.Type, "AnimeSwapPoolV1::LiquidityPool") {
			continue
		}
		tag, err := types.ParseMoveStructTag(resource.Type)
		if err != nil {
			// todo handle error
			continue
		}
		if len(tag.TypeParams) < 2 {
			continue
		}
		xTag := tag.TypeParams[0].StructTag
		yTag := tag.TypeParams[1].StructTag
		if nil == xTag || nil == yTag {
			continue
		}
		xCoinInfo, bx := p.coinListClient.GetCoinInfoByType(xTag)
		yCoinInfo, by := p.coinListClient.GetCoinInfoByType(yTag)
		if !bx || !by {
			continue
		}
		pool := NewAnimeTradingPool(p.ownerAddress, xCoinInfo, yCoinInfo, tag, resource)
		if pool == nil {
			// todo handle error
			continue
		}

		poolList = append(poolList, pool)
	}

	return poolList
}
