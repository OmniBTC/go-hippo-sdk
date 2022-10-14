package pontem

import (
	"math/big"
	"strings"

	"github.com/coming-chat/go-aptos/aptosclient"
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
	lpTag           string // todo structTag
	poolResourceTag string
}

type PoolProvider struct {
	client         *aptosclient.RestClient
	ownerAddress   string
	coinListClient *coinlist.CoinListClient
}

func NewTradingPool() base.TradingPool {
	return &TradingPool{}
}

func NewTradingPoolProvider(client *aptosclient.RestClient, ownerAddress string, coinListClient *coinlist.CoinListClient) base.TradingPoolProvider {
	return &PoolProvider{
		client:       client,
		ownerAddress: ownerAddress,
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
	reserveInAmt := t.pontemPool.CoinXReserve
	reserveOutAmt := t.pontemPool.CoinYReserve
	if !isXToY {
		inputTokenInfo, outputTokenInfo = outputTokenInfo, inputTokenInfo
		reserveInAmt, reserveOutAmt = reserveOutAmt, reserveInAmt
	}

	coinOutAmt := getCoinOutWithFees(inputAmount, reserveInAmt, reserveOutAmt)

	return base.QuoteType{
		InputSymbol:  inputTokenInfo.Symbol,
		OutputSymbol: outputTokenInfo.Symbol,
		InputAmount:  inputAmount,
		OutputAmount: coinOutAmt,
		AvgPrice: big.NewInt(0).Div(
			big.NewInt(0).Div(
				coinOutAmt,
				big.NewInt(0).Exp(
					big.NewInt(10),
					big.NewInt(int64(outputTokenInfo.Decimals)),
					nil,
				),
			),
			big.NewInt(0).Div(
				inputAmount,
				big.NewInt(0).Exp(
					big.NewInt(10),
					big.NewInt(int64(inputTokenInfo.Decimals)),
					nil,
				),
			),
		),
	}
}

func (t *TradingPool) MakePayload(input base.TokenAmount, minOut base.TokenAmount) types.EntryFunctionPayload {
	panic("not implemented")
}

/** implement base.TradingPoolProvider */

func (p *PoolProvider) LoadPoolList() []base.TradingPool {
	poolList := make([]base.TradingPool, 0)
	resources, err := p.client.GetAccountResources(p.ownerAddress)
	if err != nil {
		return poolList
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
		xCoinInfo, bx := p.coinListClient.GetCoinInfoByType(types.TokenType{
			StructTag: *xTag,
		})
		yCoinInfo, by := p.coinListClient.GetCoinInfoByType(types.TokenType{
			StructTag: *yTag,
		})
		if !bx || !by {
			continue
		}

		x := resource.Data["coin_x_reserve"].(map[string]interface{})["value"].(string)
		y := resource.Data["coin_y_reserve"].(map[string]interface{})["value"].(string)
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
			lpTag:        lpTag.GetFullName(),
		})
	}

	return poolList
}

func getCoinOutWithFees(coinInVal *big.Int, reserveInSize *big.Int, reserveOutSize *big.Int) *big.Int {
	// todo Uncorellated  use go-aptos-liquidswap
	feePct := big.NewInt(3)
	feeScale := big.NewInt(1000)
	feeMultiplier := big.NewInt(0).Sub(feeScale, feePct)
	coinInAfterFees := big.NewInt(0).Mul(coinInVal, feeMultiplier)
	newReservesInSize := big.NewInt(0).Add(
		big.NewInt(0).Mul(
			reserveInSize,
			feeScale,
		),
		coinInAfterFees,
	)

	return big.NewInt(0).Div(
		big.NewInt(0).Mul(
			coinInAfterFees,
			reserveOutSize,
		),
		newReservesInSize,
	)
}
