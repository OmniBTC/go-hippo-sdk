package pontem

import (
	"math/big"
	"strings"

	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/omnibtc/go-aptos-liquidswap/liquidswap"
	"github.com/omnibtc/go-hippo-sdk/aggregator/base"
	"github.com/omnibtc/go-hippo-sdk/aggregator/coinlist"
	"github.com/omnibtc/go-hippo-sdk/types"
)

type RawPontemPool struct {
	CoinXReserve *big.Int
	CoinYReserve *big.Int
}

type TradingPool struct {
	pontemPool      RawPontemPool
	xCoinInfo       types.CoinInfo
	yCoinInfo       types.CoinInfo
	ownerAddress    string
	lpTag           types.StructTag
	poolResourceTag string
}

type PoolProvider struct {
	client         *aptosclient.RestClient
	ownerAddress   string
	coinListClient *coinlist.CoinListClient
	resourceTypes  []string
}

func NewTradingPool() base.TradingPool {
	return &TradingPool{}
}

func NewPoolProvider(client *aptosclient.RestClient, ownerAddress string, coinListClient *coinlist.CoinListClient) base.TradingPoolProvider {
	return &PoolProvider{
		client:         client,
		ownerAddress:   ownerAddress,
		coinListClient: coinListClient,
	}
}

/** implement base.TradingPool */

func (t *TradingPool) DexType() base.DexType {
	return base.Pontem
}

func (t *TradingPool) PoolType() base.PoolType {
	return 0
}

func (t *TradingPool) IsRoutable() bool {
	return true
}

func (t *TradingPool) XCoinInfo() types.CoinInfo {
	return t.xCoinInfo
}

func (t *TradingPool) YCoinInfo() types.CoinInfo {
	return t.yCoinInfo
}

func (t *TradingPool) IsStateLoaded() bool {
	return true
}

func (t *TradingPool) GetTagE() types.TokenType {
	return &t.lpTag
}

// func (t *TradingPool) ReloadState() error {
// 	// todo 使用 client 请求 pool 数据
// 	return nil
// }

func (t *TradingPool) GetPrice() base.PriceType {
	panic("not implemented")
}

func (t *TradingPool) GetQuote(inputAmount base.TokenAmount, isXToY bool) base.QuoteType {
	if t.pontemPool.CoinXReserve == nil || t.pontemPool.CoinYReserve == nil {
		panic("pontem pool not loaded")
	}
	inputTokenInfo := t.xCoinInfo
	outputTokenInfo := t.yCoinInfo
	pool := liquidswap.PoolResource{
		CoinXReserve: t.pontemPool.CoinXReserve,
		CoinYReserve: t.pontemPool.CoinYReserve,
	}
	if !isXToY {
		inputTokenInfo, outputTokenInfo = outputTokenInfo, inputTokenInfo
		// reserve is same to fromCoin
		pool.CoinXReserve, pool.CoinYReserve = pool.CoinYReserve, pool.CoinXReserve
	}

	fromCoin := liquidswap.Coin{
		Decimals: inputTokenInfo.Decimals,
		Symbol:   inputTokenInfo.Symbol,
		Name:     inputTokenInfo.Name,
	}
	toCoin := liquidswap.Coin{
		Decimals: outputTokenInfo.Decimals,
		Symbol:   outputTokenInfo.Symbol,
		Name:     outputTokenInfo.Name,
	}
	if t.lpTag.Name == "Uncorrelated" {
		pool.CurveType = liquidswap.Uncorellated
	} else {
		pool.CurveType = liquidswap.StableCurve
	}

	// pool x y reserve should order by symbol, not same as fromcoin-tocoin
	if !liquidswap.IsSortedSymbols(fromCoin.Symbol, toCoin.Symbol) {
		pool.CoinXReserve, pool.CoinYReserve = pool.CoinYReserve, pool.CoinXReserve
	}

	coinOutAmt := liquidswap.GetAmountOut(fromCoin, toCoin, inputAmount, pool)

	return base.QuoteType{
		InputSymbol:  inputTokenInfo.Symbol,
		OutputSymbol: outputTokenInfo.Symbol,
		InputAmount:  inputAmount,
		OutputAmount: coinOutAmt,
	}
}

func (t *TradingPool) MakePayload(input base.TokenAmount, minOut base.TokenAmount) types.EntryFunctionPayload {
	panic("not implemented")
}

/** implement base.TradingPoolProvider */

func (p *PoolProvider) SetResourceTypes(resourceTypes []string) {
	if len(resourceTypes) == 0 {
		return
	}
	p.resourceTypes = resourceTypes
}

func (p *PoolProvider) LoadPoolList() []base.TradingPool {
	poolList := make([]base.TradingPool, 0)
	resources, err := p.client.GetAccountResources(p.ownerAddress)
	if err != nil {
		for _, resourceType := range p.resourceTypes {
			resource, err := p.client.GetAccountResource(p.ownerAddress, resourceType, 0)
			if err != nil {
				continue
			}
			resources = append(resources, *resource)
		}
	}
	for _, resource := range resources {
		if !strings.Contains(resource.Type, "liquidity_pool::LiquidityPool") {
			continue
		}
		tag, err := types.ParseMoveStructTag(resource.Type)
		if err != nil {
			// todo handle error
			continue
		}
		if len(tag.TypeParams) < 3 {
			continue
		}
		xTag := tag.TypeParams[0].StructTag
		yTag := tag.TypeParams[1].StructTag
		lpTag := tag.TypeParams[2].StructTag
		if nil == xTag || nil == yTag || nil == lpTag {
			continue
		}

		xCoinInfo, bx := p.coinListClient.GetCoinInfoByType(xTag)
		yCoinInfo, by := p.coinListClient.GetCoinInfoByType(yTag)
		if !bx || !by {
			continue
		}

		x := resource.Data["coin_x_reserve"].(map[string]interface{})["value"].(string)
		y := resource.Data["coin_y_reserve"].(map[string]interface{})["value"].(string)
		if x == "0" || y == "0" {
			continue
		}
		xint, b := big.NewInt(0).SetString(x, 10)
		if !b {
			continue
		}
		yint, b := big.NewInt(0).SetString(y, 10)
		if !b {
			continue
		}

		poolList = append(poolList, &TradingPool{
			pontemPool: RawPontemPool{
				CoinXReserve: xint,
				CoinYReserve: yint,
			},
			xCoinInfo:    xCoinInfo,
			yCoinInfo:    yCoinInfo,
			ownerAddress: p.ownerAddress,
			lpTag:        *lpTag,
		})
	}

	return poolList
}
