package obric

import (
	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/coming-chat/go-aptos/aptostypes"
	"github.com/omnibtc/go-hippo-sdk/aggregator/base"
	"github.com/omnibtc/go-hippo-sdk/aggregator/coinlist"
	"github.com/omnibtc/go-hippo-sdk/types"
	"math/big"
	"strings"
)

type PieceSwapPoolInfo struct {
	ReserveX                    types.Coin
	ReserveY                    types.Coin
	K                           *big.Int
	K2                          *big.Int
	Xa                          *big.Int
	Xb                          *big.Int
	M                           *big.Int
	N                           *big.Int
	XDeciMult                   *big.Int
	YDeciMult                   *big.Int
	SwapFeePerMillion           *big.Int
	ProtocolFeeSharePerThousand *big.Int
	ProtocolFeeX                types.Coin
	ProtocolFeeY                types.Coin
	TypeTag                     types.StructTag
}

func NewPieceSwapPoolInfo(resource aptostypes.AccountResource) *PieceSwapPoolInfo {
	data := resource.Data
	tag, _ := types.ParseMoveStructTag(resource.Type)
	k, b := big.NewInt(0).SetString(data["K"].(string), 10)
	if !b {
		return nil
	}
	k2, b := big.NewInt(0).SetString(data["K2"].(string), 10)
	if !b {
		return nil
	}
	xa, b := big.NewInt(0).SetString(data["Xa"].(string), 10)
	if !b {
		return nil
	}

	xb, b := big.NewInt(0).SetString(data["Xb"].(string), 10)
	if !b {
		return nil
	}
	m, b := big.NewInt(0).SetString(data["m"].(string), 10)
	if !b {
		return nil
	}
	n, b := big.NewInt(0).SetString(data["n"].(string), 10)
	if !b {
		return nil
	}
	protocolFeeSharePerThousand, b := big.NewInt(0).SetString(data["protocol_fee_share_per_thousand"].(string), 10)
	if !b {
		return nil
	}
	swapFeePerMillion, b := big.NewInt(0).SetString(data["swap_fee_per_million"].(string), 10)
	if !b {
		return nil
	}
	xDeciMult, b := big.NewInt(0).SetString(data["x_deci_mult"].(string), 10)
	if !b {
		return nil
	}
	yDeciMult, b := big.NewInt(0).SetString(data["y_deci_mult"].(string), 10)
	if !b {
		return nil
	}
	coinXValue, b := big.NewInt(0).SetString(data["reserve_x"].(map[string]interface{})["value"].(string), 10)
	if !b {
		return nil
	}
	reserveX := types.Coin{
		Value: coinXValue,
	}

	coinYValue, b := big.NewInt(0).SetString(data["reserve_y"].(map[string]interface{})["value"].(string), 10)
	if !b {
		return nil
	}
	reserveY := types.Coin{
		Value: coinYValue,
	}

	feeXValue, b := big.NewInt(0).SetString(data["protocol_fee_x"].(map[string]interface{})["value"].(string), 10)
	if !b {
		return nil
	}
	protocolFeeX := types.Coin{
		Value: feeXValue,
	}

	feeYValue, b := big.NewInt(0).SetString(data["protocol_fee_y"].(map[string]interface{})["value"].(string), 10)
	if !b {
		return nil
	}
	protocolFeeY := types.Coin{
		Value: feeYValue,
	}

	return &PieceSwapPoolInfo{
		K:                           k,
		K2:                          k2,
		Xa:                          xa,
		Xb:                          xb,
		M:                           m,
		N:                           n,
		TypeTag:                     tag,
		ReserveX:                    reserveX,
		ReserveY:                    reserveY,
		XDeciMult:                   xDeciMult,
		YDeciMult:                   yDeciMult,
		ProtocolFeeX:                protocolFeeX,
		ProtocolFeeY:                protocolFeeY,
		SwapFeePerMillion:           swapFeePerMillion,
		ProtocolFeeSharePerThousand: protocolFeeSharePerThousand,
	}
}

func (p *PieceSwapPoolInfo) quoteXToYAfterFees(amountXIn *big.Int) *big.Int {
	actualOutY := p.quoteXToY(amountXIn)
	totalFees := new(big.Int).Div(new(big.Int).Mul(actualOutY, p.SwapFeePerMillion), big.NewInt(1000000))
	return new(big.Int).Sub(actualOutY, totalFees)
}

func (p *PieceSwapPoolInfo) quoteYToXAfterFees(amountYIn *big.Int) *big.Int {
	actualOutX := p.quoteYTox(amountYIn)
	totalFees := new(big.Int).Div(new(big.Int).Mul(actualOutX, p.SwapFeePerMillion), big.NewInt(1000000))
	return new(big.Int).Sub(actualOutX, totalFees)
}

func (p *PieceSwapPoolInfo) quoteXToY(amountXIn *big.Int) *big.Int {
	currentX := new(big.Int).Mul(p.ReserveX.Value, p.XDeciMult)
	currentY := new(big.Int).Mul(p.ReserveY.Value, p.YDeciMult)
	inputX := new(big.Int).Mul(amountXIn, p.XDeciMult)
	optOutPutY := getSwapXToYOut(currentX, currentY, inputX, p.K, p.K2, p.Xa, p.Xb, p.M, p.N)
	return new(big.Int).Div(optOutPutY, p.YDeciMult)
}

func (p *PieceSwapPoolInfo) quoteYTox(amountYIn *big.Int) *big.Int {
	currentX := new(big.Int).Mul(p.ReserveX.Value, p.XDeciMult)
	currentY := new(big.Int).Mul(p.ReserveY.Value, p.YDeciMult)
	inputY := new(big.Int).Mul(amountYIn, p.YDeciMult)
	optOutPutX := getSwapYToXOut(currentX, currentY, inputY, p.K, p.K2, p.Xa, p.Xb, p.M, p.N)
	return new(big.Int).Div(optOutPutX, p.XDeciMult)
}

type ObricTradingPool struct {
	pool      *PieceSwapPoolInfo
	xCoinInfo types.CoinInfo
	yCoinInfo types.CoinInfo
}

func NewObricTradingPool(xCoinInfo, yCoinInfo types.CoinInfo, resource aptostypes.AccountResource) base.TradingPool {
	pool := NewPieceSwapPoolInfo(resource)
	if pool == nil {
		return nil
	}
	return &ObricTradingPool{
		pool:      pool,
		xCoinInfo: xCoinInfo,
		yCoinInfo: yCoinInfo,
	}
}

func (t *ObricTradingPool) DexType() base.DexType {
	return base.Obric
}

func (t *ObricTradingPool) PoolType() base.PoolType {
	return 0
}

func (t *ObricTradingPool) IsRoutable() bool {
	return true
}

func (t *ObricTradingPool) XCoinInfo() types.CoinInfo {
	return t.xCoinInfo
}

func (t *ObricTradingPool) YCoinInfo() types.CoinInfo {
	return t.yCoinInfo
}

func (t *ObricTradingPool) IsStateLoaded() bool {
	return t.pool != nil
}

// ReloadState() error
func (t *ObricTradingPool) GetPrice() base.PriceType {
	panic("not implemented")
}

func (t *ObricTradingPool) GetQuote(inputAmount base.TokenAmount, isXToY bool) base.QuoteType {
	if !t.IsStateLoaded() {
		panic("pancake pool not loaded")
	}

	inputTokenInfo := t.xCoinInfo
	outputTokenInfo := t.yCoinInfo
	if !isXToY {
		inputTokenInfo, outputTokenInfo = outputTokenInfo, inputTokenInfo
	}

	var outputAmount *big.Int
	if isXToY {
		outputAmount = t.pool.quoteXToYAfterFees(inputAmount)
	} else {
		outputAmount = t.pool.quoteYToXAfterFees(inputAmount)
	}

	return base.QuoteType{
		InputSymbol:  inputTokenInfo.Symbol,
		OutputSymbol: outputTokenInfo.Symbol,
		InputAmount:  inputAmount,
		OutputAmount: outputAmount,
	}
}

func (t *ObricTradingPool) GetTagE() types.TokenType {
	return types.U8
}

func (t *ObricTradingPool) MakePayload(input base.TokenAmount, minOut base.TokenAmount, isXToY bool) types.EntryFunctionPayload {
	panic("not implemented")
}

type ObricPoolProvider struct {
	client         *aptosclient.RestClient
	ownerAddress   string
	coinListClient *coinlist.CoinListClient
	resourceTypes  []string
}

func NewPoolProvider(client *aptosclient.RestClient, ownerAddress string, coinListClient *coinlist.CoinListClient) base.TradingPoolProvider {
	return &ObricPoolProvider{
		client:         client,
		ownerAddress:   ownerAddress,
		coinListClient: coinListClient,
	}
}

func (p *ObricPoolProvider) SetResourceTypes(resourceTypes []string) {
	if len(resourceTypes) == 0 {
		return
	}
	p.resourceTypes = resourceTypes
}

func (p *ObricPoolProvider) LoadPoolList() []base.TradingPool {
	poolList := make([]base.TradingPool, 0)
	resources, err := p.client.GetAccountResources(p.ownerAddress, 0)
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
		if !strings.Contains(resource.Type, "piece_swap::PieceSwapPoolInfo") {
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
		pool := NewObricTradingPool(xCoinInfo, yCoinInfo, resource)
		if pool == nil {
			// todo handle error
			continue
		}

		poolList = append(poolList, pool)
	}

	return poolList
}
