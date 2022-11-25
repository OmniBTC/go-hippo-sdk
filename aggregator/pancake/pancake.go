package pancake

import (
	"fmt"
	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/coming-chat/go-aptos/aptostypes"
	"github.com/omnibtc/go-hippo-sdk/aggregator/base"
	"github.com/omnibtc/go-hippo-sdk/aggregator/coinlist"
	"github.com/omnibtc/go-hippo-sdk/types"
	"math/big"
	"strings"
)

type Pool struct {
	reserveX           *big.Int
	reserveY           *big.Int
	blockTimestampLast *big.Int
}

func NewPool(resource aptostypes.AccountResource) *Pool {
	data := resource.Data
	blockTimestampLast, b := big.NewInt(0).SetString(data["block_timestamp_last"].(string), 10)
	if !b {
		return nil
	}
	reserveX, b := big.NewInt(0).SetString(data["reserve_x"].(string), 10)
	if !b {
		return nil
	}
	reserveY, b := big.NewInt(0).SetString(data["reserve_y"].(string), 10)
	if !b {
		return nil
	}
	if reserveX.Cmp(big.NewInt(0)) == 0 || reserveY.Cmp(big.NewInt(0)) == 0 {
		return nil
	}
	return &Pool{
		reserveX:           reserveX,
		reserveY:           reserveY,
		blockTimestampLast: blockTimestampLast,
	}
}

func (p *Pool) tokenReserves() (reserveX, reserveY, blockTimestampLast *big.Int) {
	return p.reserveX, p.reserveY, p.blockTimestampLast
}

type TradingPool struct {
	pool          *Pool
	xCoinInfo     types.CoinInfo
	yCoinInfo     types.CoinInfo
	owner         string
	scriptAddress string
}

func NewTradingPool(xCoinInfo, yCoinInfo types.CoinInfo, owner string, resource aptostypes.AccountResource, scriptAddress string) base.TradingPool {
	pool := NewPool(resource)
	if pool == nil {
		return nil
	}
	return &TradingPool{
		pool:          pool,
		xCoinInfo:     xCoinInfo,
		yCoinInfo:     yCoinInfo,
		owner:         owner,
		scriptAddress: scriptAddress,
	}
}

func (t *TradingPool) DexType() base.DexType {
	return base.Pancake
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
	return t.pool != nil
}

// ReloadState() error
func (t *TradingPool) GetPrice() base.PriceType {
	panic("not implemented")
}

func (t *TradingPool) GetQuote(inputAmount base.TokenAmount, isXToY bool) base.QuoteType {
	if !t.IsStateLoaded() {
		panic("pancake pool not loaded")
	}

	inputTokenInfo := t.xCoinInfo
	outputTokenInfo := t.yCoinInfo
	rin, rout, _ := t.pool.tokenReserves()
	if !isXToY {
		inputTokenInfo, outputTokenInfo = outputTokenInfo, inputTokenInfo
		rin, rout = rout, rin
	}

	coinOutAmt := getAmountOut(inputAmount, rin, rout)

	return base.QuoteType{
		InputSymbol:  inputTokenInfo.Symbol,
		OutputSymbol: outputTokenInfo.Symbol,
		InputAmount:  inputAmount,
		OutputAmount: coinOutAmt,
	}
}

func (t *TradingPool) GetTagE() types.TokenType {
	return types.U8
}

func (t *TradingPool) MakePayload(input base.TokenAmount, minOut base.TokenAmount, isXToY bool) types.EntryFunctionPayload {
	xTokenType := t.xCoinInfo.TokenType
	yTokenType := t.yCoinInfo.TokenType
	if !isXToY {
		xTokenType, yTokenType = yTokenType, xTokenType
	}

	typeArgs := make([]string, 0)
	typeArgs = append(typeArgs, xTokenType.GetFullName(), yTokenType.GetFullName())
	return types.EntryFunctionPayload{
		Function: fmt.Sprintf("%s::%s::%s", t.scriptAddress, "router", "swap_exact_input"),
		TypeArgs: typeArgs,
		Args: []interface{}{
			input,
			minOut,
		},
	}
}

func getAmountOut(amountIn, reserveIn, reserveOut *big.Int) *big.Int {
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
	amountInWithFee := new(big.Int).Mul(amountIn, big.NewInt(9975))
	numerator := new(big.Int).Mul(amountInWithFee, reserveOut)
	denominator := new(big.Int).Add(new(big.Int).Mul(reserveIn, big.NewInt(10000)), amountInWithFee)
	amountOut := new(big.Int).Div(numerator, denominator)
	return amountOut
}

type PancakePoolProvider struct {
	client         *aptosclient.RestClient
	ownerAddress   string
	coinListClient *coinlist.CoinListClient
	resourceTypes  []string
	scriptAddress  string
}

func NewPoolProvider(client *aptosclient.RestClient, ownerAddress string, coinListClient *coinlist.CoinListClient, scriptAddress string) base.TradingPoolProvider {
	return &PancakePoolProvider{
		client:         client,
		ownerAddress:   ownerAddress,
		coinListClient: coinListClient,
		scriptAddress:  scriptAddress,
	}
}

func (p *PancakePoolProvider) SetResourceTypes(resourceTypes []string) {
	if len(resourceTypes) == 0 {
		return
	}
	p.resourceTypes = resourceTypes
}

func (p *PancakePoolProvider) LoadPoolList() []base.TradingPool {
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
		if !strings.Contains(resource.Type, "swap::TokenPairReserve") {
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
		pool := NewTradingPool(xCoinInfo, yCoinInfo, p.ownerAddress, resource, p.scriptAddress)
		if pool == nil {
			// todo handle error
			continue
		}

		poolList = append(poolList, pool)
	}

	return poolList
}
