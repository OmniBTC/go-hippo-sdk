package base

import (
	"fmt"
	"math/big"

	"github.com/omnibtc/go-hippo-sdk/types"
)

type DexType int

const (
	_ DexType = iota
	Hippo
	Econia
	Pontem
	Basiq
	Ditto
	Tortuga
	Aptosswap
	Aux // 8
	AnimeSwap
	Cetus
	Pancake
	Obric
)

type PoolType uint64

type TokenAmount *big.Int
type TokenAmountRatio *big.Int

type PriceType struct {
	XToY TokenAmountRatio
	YToX TokenAmountRatio
}

type QuoteType struct {
	InputSymbol  string
	OutputSymbol string
	InputAmount  TokenAmount // bigint, eg. 100000000, = amount * 10^decimals
	OutputAmount TokenAmount
}

type TradingPool interface {
	DexType() DexType
	PoolType() PoolType
	IsRoutable() bool
	XCoinInfo() types.CoinInfo
	YCoinInfo() types.CoinInfo
	IsStateLoaded() bool
	// ReloadState() error
	GetPrice() PriceType
	GetQuote(inputAmount TokenAmount, isXToY bool) QuoteType
	GetTagE() types.TokenType
	MakePayload(input TokenAmount, minOut TokenAmount, isXToY bool) types.EntryFunctionPayload
}

type TradingPoolProvider interface {
	LoadPoolList() []TradingPool
	SetResourceTypes(resourceTypes []string)
}

// TradeStep is a single trade step involving a Pool and a direction (X-to-Y or Y-to-X)
type TradeStep struct {
	Pool   TradingPool
	IsXtoY bool
}

func NewTradeStep(pool TradingPool, isXtoY bool) TradeStep {
	return TradeStep{
		Pool:   pool,
		IsXtoY: isXtoY,
	}
}

func (ts *TradeStep) XCoinInfo() types.CoinInfo {
	if ts.IsXtoY {
		return ts.Pool.XCoinInfo()
	} else {
		return ts.Pool.YCoinInfo()
	}
}

func (ts *TradeStep) YCoinInfo() types.CoinInfo {
	if ts.IsXtoY {
		return ts.Pool.YCoinInfo()
	} else {
		return ts.Pool.XCoinInfo()
	}
}

func (ts *TradeStep) XTag() {
	//c := ts.XCoinInfo()
	// todo
	panic("todo")
}

func (ts *TradeStep) YTag() {
	// c := ts.YCoinInfo()
	// todo
	panic("todo")
}

func (ts *TradeStep) GetPrice() PriceType {
	price := ts.Pool.GetPrice()
	if ts.IsXtoY {
		return price
	} else {
		return PriceType{
			XToY: price.YToX,
			YToX: price.XToY,
		}
	}
}

func (ts *TradeStep) GetQuote(inputAmount TokenAmount) QuoteType {
	return ts.Pool.GetQuote(inputAmount, ts.IsXtoY)
}

func (ts *TradeStep) GetTagE() types.TokenType {
	return ts.Pool.GetTagE()
}

type TradeRoute struct {
	Tokens []types.CoinInfo
	Steps  []TradeStep
}

type RouteAndQuote struct {
	Route TradeRoute
	Quote *QuoteType
}

func NewTradeRoute(steps []TradeStep) TradeRoute {
	if len(steps) < 1 {
		panic("route need at least on trade step")
	}
	tr := TradeRoute{
		Tokens: make([]types.CoinInfo, 0),
		Steps:  steps,
	}
	tokenFullName := steps[0].XCoinInfo().TokenType.GetFullName()
	tr.Tokens = append(tr.Tokens, steps[0].XCoinInfo())
	for _, step := range steps {
		xFullName := step.XCoinInfo().TokenType.GetFullName()
		yFullName := step.YCoinInfo().TokenType.GetFullName()
		if xFullName != tokenFullName {
			panic(fmt.Errorf("mismatching tokens in route, expect %s but received %s", tokenFullName, xFullName))
		}
		tokenFullName = yFullName
		tr.Tokens = append(tr.Tokens, step.YCoinInfo())
	}
	return tr
}

func (tr *TradeRoute) XCoinInfo() types.CoinInfo {
	return tr.Steps[0].XCoinInfo()
}

func (tr *TradeRoute) YCoinInfo() types.CoinInfo {
	return tr.Steps[len(tr.Steps)-1].YCoinInfo()
}

func (tr *TradeRoute) XTag() string {
	return tr.XCoinInfo().TokenType.GetFullName()
}

func (tr *TradeRoute) YTag() string {
	return tr.YCoinInfo().TokenType.GetFullName()
}

func (tr *TradeRoute) GetPrice() PriceType {
	xToy := big.NewInt(1)
	yTox := big.NewInt(1)
	for _, step := range tr.Steps {
		price := step.Pool.GetPrice()
		xToy = big.NewInt(0).Mul(xToy, price.XToY)
		yTox = big.NewInt(0).Mul(yTox, price.YToX)
	}
	return PriceType{
		XToY: xToy,
		YToX: yTox,
	}
}

func (tr *TradeRoute) GetQuote(inputAmount TokenAmount) *QuoteType {
	outputAmount := inputAmount
	for _, step := range tr.Steps {
		outputAmount = step.GetQuote(outputAmount).OutputAmount
	}
	return &QuoteType{
		InputSymbol:  tr.XCoinInfo().Symbol,
		OutputSymbol: tr.YCoinInfo().Symbol,
		InputAmount:  inputAmount,
		OutputAmount: outputAmount,
	}
}

func (tr *TradeRoute) HasRoundTrip() bool {
	s := make(map[string]struct{})
	for _, token := range tr.Tokens {
		if _, ok := s[token.TokenType.GetFullName()]; ok {
			return true
		} else {
			s[token.TokenType.GetFullName()] = struct{}{}
		}
	}
	return false
}

// TryMakeRawPayload return raw router payload when step length is 1
// raw router payload will cost less gas then hippo on_step_route
func (tr *TradeRoute) TryMakeRawPayload(inputAmount, minOutAmount *big.Int) (types.EntryFunctionPayload, bool) {
	if len(tr.Steps) > 1 {
		return types.EntryFunctionPayload{}, false
	}
	switch tr.Steps[0].Pool.DexType() {
	case Aux, Pancake, Pontem:
		return tr.Steps[0].Pool.MakePayload(inputAmount, minOutAmount, tr.Steps[0].IsXtoY), true
	default:
		return types.EntryFunctionPayload{}, false
	}
}

func (tr *TradeRoute) MakePayload(inputAmount, minOutAmount *big.Int) types.EntryFunctionPayload {
	inputAmountU64 := inputAmount.Uint64()
	minOutAmountU64 := minOutAmount.Uint64()
	switch len(tr.Steps) {
	case 1:
		step0 := tr.Steps[0]
		return types.BuildPayloadOneStepRoute(
			uint8(step0.Pool.DexType()),
			uint64(step0.Pool.PoolType()),
			step0.IsXtoY,
			inputAmountU64,
			minOutAmountU64,
			[]types.TokenType{tr.XCoinInfo().TokenType, tr.YCoinInfo().TokenType, step0.GetTagE()},
		)
	case 2:
		step0 := tr.Steps[0]
		step1 := tr.Steps[1]
		return types.BuildPayloadTwoStepRoute(
			uint8(step0.Pool.DexType()),
			uint64(step0.Pool.PoolType()),
			step0.IsXtoY,
			uint8(step1.Pool.DexType()),
			uint64(step1.Pool.PoolType()),
			step1.IsXtoY,
			inputAmountU64,
			minOutAmountU64,
			[]types.TokenType{
				tr.Tokens[0].TokenType,
				tr.Tokens[1].TokenType,
				tr.Tokens[2].TokenType,
				step0.GetTagE(),
				step1.GetTagE(),
			}, // X, Y, Z, E1, E2
		)
	case 3:
		step0 := tr.Steps[0]
		step1 := tr.Steps[1]
		step2 := tr.Steps[2]
		return types.BuildPayloadThreeStepRoute(
			uint8(step0.Pool.DexType()),
			uint64(step0.Pool.PoolType()),
			step0.IsXtoY,
			uint8(step1.Pool.DexType()),
			uint64(step1.Pool.PoolType()),
			step1.IsXtoY,
			uint8(step2.Pool.DexType()),
			uint64(step2.Pool.PoolType()),
			step2.IsXtoY,
			inputAmountU64,
			minOutAmountU64,
			[]types.TokenType{
				tr.Tokens[0].TokenType,
				tr.Tokens[1].TokenType,
				tr.Tokens[2].TokenType,
				tr.Tokens[3].TokenType,
				step0.GetTagE(),
				step1.GetTagE(),
				step2.GetTagE(),
			},
		)
	default:
		panic("unreachable")
	}
}

func (t DexType) Name() string {
	return DexTypeName(t)
}

func DexTypeName(t DexType) string {
	switch t {
	case Hippo:
		return "Hippo"
	case Econia:
		return "Econia"
	case Pontem:
		return "Pontem"
	case Basiq:
		return "Basiq"
	case Ditto:
		return "Ditto"
	case Tortuga:
		return "Tortuga"
	case Aptosswap:
		return "Aptosswap"
	case Aux:
		return "Aux"
	case AnimeSwap:
		return "AnimeSwap"
	case Cetus:
		return "Cetus"
	case Pancake:
		return "Pancake"
	case Obric:
		return "Obric"
	}
	return ""
}

func BigIntToUint64(x, y *big.Int) (uint64, uint64) {
	var _x, _y uint64
	if x == nil {
		_x = 0
	} else {
		_x = x.Uint64()
	}
	if y == nil {
		_y = 0
	} else {
		_y = y.Uint64()
	}
	return _x, _y
}

// func ReloadAllPool(pools []TradingPool) {
// 	wg := sync.WaitGroup{}
// 	for _, p := range pools {
// 		wg.Add(1)
// 		go func(p TradingPool) {
// 			defer wg.Done()
// 			p.ReloadState()
// 		}(p)
// 	}
// 	wg.Wait()
// }
