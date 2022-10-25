package basiq

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/omnibtc/go-hippo-sdk/aggregator/base"
	"github.com/omnibtc/go-hippo-sdk/aggregator/coinlist"
	"github.com/omnibtc/go-hippo-sdk/types"
)

type TradingPool struct {
	xCoinInfo    types.CoinInfo
	yCoinInfo    types.CoinInfo
	ownerAddress string

	// pool info on net
	feeBips            int
	rebateBips         int
	coinXReserve       *big.Int
	coinYReserve       *big.Int
	xDecimalAdjustment *big.Int
	yDecimalAdjustment *big.Int
	xPrice             *big.Int
	yPrice             *big.Int
}

func NewTradingPool() base.TradingPool {
	return &TradingPool{}
}

func (t *TradingPool) DexType() base.DexType {
	return base.Basiq
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
	return t.coinXReserve != nil && t.coinYReserve != nil
}

// ReloadState() error
func (t *TradingPool) GetPrice() base.PriceType {
	panic("not implemented") // TODO: Implement
}

func (t *TradingPool) GetQuote(inputAmount base.TokenAmount, isXToY bool) base.QuoteType {
	if !t.IsStateLoaded() {
		panic("state not loaded")
	}
	inputTokenInfo := t.xCoinInfo
	outputTokenInfo := t.yCoinInfo
	reserveInAmt := t.coinXReserve
	reserveOutAmt := t.coinYReserve
	coinInAdjust := t.xDecimalAdjustment
	coinOutAdjust := t.yDecimalAdjustment
	coinInPrice := t.xPrice
	coinOutPrice := t.yPrice
	if !isXToY {
		inputTokenInfo, outputTokenInfo = outputTokenInfo, inputTokenInfo
		reserveInAmt, reserveOutAmt = reserveOutAmt, reserveInAmt
		coinInAdjust, coinOutAdjust = coinOutAdjust, coinInAdjust
		coinInPrice, coinOutPrice = coinOutPrice, coinInPrice
	}

	feeBip := big.NewInt(int64(t.feeBips))
	rebateBips := big.NewInt(int64(t.rebateBips))
	coinOutAmount := calcSwapOutput(
		big.NewInt(0).Mul(inputAmount, coinInAdjust),
		big.NewInt(0).Mul(reserveInAmt, coinInAdjust),
		big.NewInt(0).Mul(reserveOutAmt, coinOutAdjust),
		coinInPrice,
		coinOutPrice,
		feeBip,
		rebateBips,
	)
	coinOutAmount = big.NewInt(0).Div(
		coinOutAmount,
		coinOutAdjust,
	)
	return base.QuoteType{
		InputSymbol:  inputTokenInfo.Symbol,
		OutputSymbol: outputTokenInfo.Symbol,
		InputAmount:  inputAmount,
		OutputAmount: coinOutAmount,
	}
}

func (t *TradingPool) GetTagE() types.TokenType {
	return types.U8
}

func (t *TradingPool) MakePayload(input base.TokenAmount, minOut base.TokenAmount) types.EntryFunctionPayload {
	panic("not implemented") // TODO: Implement
}

type BasiqPoolProvider struct {
	client         *aptosclient.RestClient
	ownerAddress   string
	coinListClient *coinlist.CoinListClient
}

func NewPoolProvider(client *aptosclient.RestClient, ownerAddress string, coinListClient *coinlist.CoinListClient) base.TradingPoolProvider {
	return &BasiqPoolProvider{
		client:         client,
		ownerAddress:   ownerAddress,
		coinListClient: coinListClient,
	}
}

func (p *BasiqPoolProvider) LoadPoolList() []base.TradingPool {
	poolList := make([]base.TradingPool, 0)
	resources, err := p.client.GetAccountResources(p.ownerAddress)
	if err != nil {
		return poolList
	}

	for _, resource := range resources {
		if !strings.Contains(resource.Type, "dex::BasiqPoolV1") {
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
		x := resource.Data["x_reserve"].(map[string]interface{})["value"].(string)
		y := resource.Data["y_reserve"].(map[string]interface{})["value"].(string)
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
		feeBips, _ := strconv.Atoi(resource.Data["fee_bips"].(string))
		rebateBips, _ := strconv.Atoi(resource.Data["rebate_bips"].(string))
		xDecimalAdjustment, _ := big.NewInt(0).SetString(resource.Data["x_decimal_adjustment"].(string), 10)
		yDecimalAdjustment, _ := big.NewInt(0).SetString(resource.Data["y_decimal_adjustment"].(string), 10)
		xPrice, _ := big.NewInt(0).SetString(resource.Data["x_price"].(string), 10)
		yPrice, _ := big.NewInt(0).SetString(resource.Data["y_price"].(string), 10)
		poolList = append(poolList, &TradingPool{
			xCoinInfo:          xCoinInfo,
			yCoinInfo:          yCoinInfo,
			feeBips:            feeBips,
			coinXReserve:       xint,
			coinYReserve:       yint,
			ownerAddress:       p.ownerAddress,
			rebateBips:         rebateBips,
			xDecimalAdjustment: xDecimalAdjustment,
			yDecimalAdjustment: yDecimalAdjustment,
			xPrice:             xPrice,
			yPrice:             yPrice,
		})
	}
	return poolList
}

func calcSwapOutput(
	inputAmount,
	inputReserve,
	outputReserve,
	inputPrice,
	outputPrice,
	feeBips,
	rebateBips *big.Int,
) *big.Int {
	fairInputVal := big.NewInt(0).Mul(inputAmount, inputPrice)
	inputReserveVal := big.NewInt(0).Mul(inputReserve, inputPrice)
	outputReserveVal := big.NewInt(0).Mul(outputReserve, outputPrice)

	preTradeImbalance := imbalanceRatio(inputReserveVal, outputReserveVal)
	postTradeImbalance := imbalanceRatio(
		big.NewInt(0).Add(inputReserveVal, fairInputVal),
		big.NewInt(0).Sub(outputReserveVal, fairInputVal),
	)

	// has rebate
	if postTradeImbalance.Cmp(preTradeImbalance) < 0 {
		// fairInputValue / outputPrice * (10000 - (feeBips - rebateBips)) / 10000
		return big.NewInt(0).Div(
			big.NewInt(0).Mul(
				big.NewInt(0).Div(
					fairInputVal,
					outputPrice,
				),
				big.NewInt(0).Sub(
					big.NewInt(10000),
					big.NewInt(0).Sub(
						feeBips,
						rebateBips,
					),
				),
			),
			big.NewInt(10000),
		)
	}

	// has penalty
	if postTradeImbalance.Cmp(big.NewInt(7500)) > 0 {
		postTradeSurplus := big.NewInt(0).Div(
			big.NewInt(0).Sub(
				postTradeImbalance,
				big.NewInt(7500),
			),
			big.NewInt(100),
		)
		penaltyBips := big.NewInt(0).Mul(
			big.NewInt(0).Exp(postTradeSurplus, big.NewInt(2), nil),
			big.NewInt(2),
		)
		return big.NewInt(0).Div(
			big.NewInt(0).Mul(
				big.NewInt(0).Div(fairInputVal, outputPrice),
				big.NewInt(0).Sub(
					big.NewInt(0).Sub(
						big.NewInt(10000),
						feeBips,
					),
					penaltyBips,
				),
			),
			big.NewInt(10000),
		)
	} else {
		return big.NewInt(0).Div(
			big.NewInt(0).Mul(
				big.NewInt(0).Div(fairInputVal, outputPrice),
				big.NewInt(0).Sub(
					big.NewInt(10000),
					feeBips,
				),
			),
			big.NewInt(10000),
		)
	}
}

func imbalanceRatio(x, y *big.Int) *big.Int {
	total := big.NewInt(0).Add(x, y)
	if x.Cmp(y) > 0 {
		return big.NewInt(0).Div(
			big.NewInt(0).Mul(x, big.NewInt(10000)),
			total,
		)
	} else {
		return big.NewInt(0).Div(
			big.NewInt(0).Mul(y, big.NewInt(10000)),
			total,
		)
	}
}
