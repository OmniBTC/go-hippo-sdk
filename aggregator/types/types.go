package types

import (
	"fmt"
	"math/big"
	"sync"
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
	AvgPrice     TokenAmountRatio
	InitialPrice TokenAmountRatio
	FinalPrice   TokenAmountRatio
	PriceImpact  int
}

type TradingPool interface {
	DexType() DexType
	PoolType() PoolType
	IsRoutable() bool
	XCoinInfo() CoinInfo
	YCoinInfo() CoinInfo
	IsStateLoaded() bool
	ReloadState(App) error
	GetPrice() PriceType
	GetQuote(TokenAmount, bool) QuoteType
	MakePayload(input TokenAmount, minOut TokenAmount) EntryFunctionPayload
}

type TradingPoolProvider interface {
	LoadPoolList() []TradingPool
}

// TradeStep is a single trade step involving a Pool and a direction (X-to-Y or Y-to-X)
type TradeStep struct {
	Pool   TradingPool
	IsXtoY bool
}

func NewTradeStep(pool TradingPool, isXtoY bool) *TradeStep {
	return &TradeStep{
		Pool:   pool,
		IsXtoY: isXtoY,
	}
}

func (ts *TradeStep) XCoinInfo() CoinInfo {
	if ts.IsXtoY {
		return ts.Pool.XCoinInfo()
	} else {
		return ts.Pool.YCoinInfo()
	}
}

func (ts *TradeStep) YCoinInfo() CoinInfo {
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

func (ts *TradeStep) GetTagE() TokenType {
	panic("todo")
}

type TradeRoute struct {
	tokens []CoinInfo
	steps  []TradeStep
}

func NewTradeRoute(steps []TradeStep) *TradeRoute {
	if len(steps) < 1 {
		panic("route need at least on trade step")
	}
	tr := &TradeRoute{
		tokens: make([]CoinInfo, 0),
		steps:  steps,
	}
	tokenFullName := steps[0].XCoinInfo().TokenType.FullName()
	tr.tokens = append(tr.tokens, steps[0].XCoinInfo())
	for _, step := range steps {
		xFullName := step.XCoinInfo().TokenType.FullName()
		yFullName := step.YCoinInfo().TokenType.FullName()
		if xFullName != tokenFullName {
			panic(fmt.Errorf("mismatching tokens in route, expect %s but received %s", tokenFullName, xFullName))
		}
		tokenFullName = yFullName
		tr.tokens = append(tr.tokens, step.YCoinInfo())
	}
	return tr
}

func (tr *TradeRoute) XCoinInfo() CoinInfo {
	return tr.steps[0].XCoinInfo()
}

func (tr *TradeRoute) YCoinInfo() CoinInfo {
	return tr.steps[len(tr.steps)-1].YCoinInfo()
}

func (tr *TradeRoute) XTag() string {
	return tr.XCoinInfo().TokenType.ToTypeTag()
}

func (tr *TradeRoute) YTag() string {
	return tr.YCoinInfo().TokenType.ToTypeTag()
}

func (tr *TradeRoute) GetPrice() PriceType {
	xToy := big.NewInt(1)
	yTox := big.NewInt(1)
	for _, step := range tr.steps {
		price := step.Pool.GetPrice()
		xToy = big.NewInt(0).Mul(xToy, price.XToY)
		yTox = big.NewInt(0).Mul(yTox, price.YToX)
	}
	return PriceType{
		XToY: xToy,
		YToX: yTox,
	}
}

func (tr *TradeRoute) GetQuote(inputAmount TokenAmount) QuoteType {
	outputAmount := inputAmount
	for _, step := range tr.steps {
		outputAmount = step.GetQuote(outputAmount).OutputAmount
	}
	return QuoteType{
		InputSymbol:  tr.XCoinInfo().TokenType.Symbol,
		OutputSymbol: tr.YCoinInfo().TokenType.Symbol,
		InputAmount:  inputAmount,
		OutputAmount: outputAmount,
		AvgPrice:     big.NewInt(0).Div(outputAmount, inputAmount),
	}
}

func (tr *TradeRoute) HasRoundTrip() bool {
	s := make(map[string]struct{})
	for _, token := range tr.tokens {
		if _, ok := s[token.TokenType.FullName()]; ok {
			return true
		} else {
			s[token.TokenType.FullName()] = struct{}{}
		}
	}
	return false
}

func (tr *TradeRoute) MakePayload(inputAmount, minOutAmount *big.Int) EntryFunctionPayload {
	inputAmountU64 := inputAmount.Uint64()
	minOutAmountU64 := minOutAmount.Uint64()
	switch len(tr.steps) {
	case 1:
		step0 := tr.steps[0]
		return buildPayloadOneStepRoute(
			uint8(step0.Pool.DexType()),
			uint64(step0.Pool.PoolType()),
			step0.IsXtoY,
			inputAmountU64,
			minOutAmountU64,
			[]TokenType{tr.XCoinInfo().TokenType, tr.YCoinInfo().TokenType, step0.GetTagE()},
		)
	case 2:
		step0 := tr.steps[0]
		step1 := tr.steps[1]
		return buildPayloadTwoStepRoute(
			uint8(step0.Pool.DexType()),
			uint64(step0.Pool.PoolType()),
			step0.IsXtoY,
			uint8(step1.Pool.DexType()),
			uint64(step1.Pool.PoolType()),
			step1.IsXtoY,
			inputAmountU64,
			minOutAmountU64,
			[]TokenType{
				tr.tokens[0].TokenType,
				tr.tokens[1].TokenType,
				tr.tokens[2].TokenType,
				step0.GetTagE(),
				step1.GetTagE(),
			}, // X, Y, Z, E1, E2
		)
	case 3:
		step0 := tr.steps[0]
		step1 := tr.steps[1]
		step2 := tr.steps[2]
		return buildPayloadThreeStepRoute(
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
			[]TokenType{
				tr.tokens[0].TokenType,
				tr.tokens[1].TokenType,
				tr.tokens[2].TokenType,
				tr.tokens[3].TokenType,
				step0.GetTagE(),
				step1.GetTagE(),
				step2.GetTagE(),
			},
		)
	default:
		panic("unreachable")
	}
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
	}
	return ""
}

func ReloadAllPool(app App, pools []TradingPool) {
	wg := sync.WaitGroup{}
	for _, p := range pools {
		wg.Add(1)
		go func(p TradingPool) {
			defer func() {
				wg.Done()
			}()
			p.ReloadState(app)
		}(p)
	}
	wg.Wait()
}
