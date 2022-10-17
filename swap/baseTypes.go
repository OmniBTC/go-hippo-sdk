package swap

import (
	"errors"
	"fmt"
	transactionbuilder "github.com/coming-chat/go-aptos/transaction_builder"
	"github.com/omnibtc/go-hippo-sdk/types"
	"github.com/shopspring/decimal"
	"math"
	"math/big"
	"strings"
)

type UITokenAmount *big.Int

type PriceType struct {
	XToY int64
	YToX int64
}

type QuoteType struct {
	InputSymbol  string
	OutputSymbol string
	InputUiAmt   UITokenAmount
	OutputUiAmt  UITokenAmount
	AvgPrice     int64
	InitialPrice int64
	FinalPrice   int64
	PriceImpact  int64
}

type PoolType int

const (
	_ PoolType = iota
	CONSTANT_PRODUCT
	STABLE_CURVE
	THREE_PIECE
)

func PoolTypeToName(poolType PoolType) (string, error) {
	var (
		name string
		err  error
	)
	switch poolType {
	case CONSTANT_PRODUCT:
		name = "ConstantProduct"
	case STABLE_CURVE:
		name = "StableCurve"
	case THREE_PIECE:
		name = "ThreePiece"
	default:
		err = errors.New(fmt.Sprintf("%d not fount pool name", poolType))
	}
	return name, err
}

type TradeRouteInterface interface {
	GetCurrentPrice() PriceType
	GetQuote(uiAmount UITokenAmount) QuoteType
	MakeSwapPayload(amountIn, minAmountOut UITokenAmount) transactionbuilder.TransactionPayloadEntryFunction
}

type TradeRoute struct {
	TradeRouteInterface
	XCoinInfo types.CoinInfo
	YCoinInfo types.CoinInfo
}

func NewTradeRoute(xCoinInfo, yCoinInfo types.CoinInfo) TradeRoute {
	return TradeRoute{
		XCoinInfo: xCoinInfo,
		YCoinInfo: yCoinInfo,
	}
}

func (t *TradeRoute) XTag() string {
	return "" // todo return XCoinInfo.TokenType
}

func (t *TradeRoute) YTag() string {
	return "" // todo return YCoinInfo.TokenType
}

type HippoPoolInterface interface {
	EstimateWithdrawalOutput(lpUiAmount, lpSupplyUiAmt UITokenAmount) (xUiAmt, yUiAmt UITokenAmount)
	EstimateNeededYFromXDeposit(xUiAmt UITokenAmount) UITokenAmount
	EstimateNeededXFromYDeposit(yUiAmt UITokenAmount) UITokenAmount
	GetPoolType() PoolType
	XUiBalance() int64
	YUiBalance() int64
	GetId() string
	GetCurrentPriceDirectional(isXtoY bool) PriceType
	GetQuoteDirectional(uiAmount UITokenAmount, isXtoY bool) QuoteType
	MakeSwapPayloadDirectional(amountIn, minAmountOut UITokenAmount, isXtoY bool) transactionbuilder.TransactionPayloadEntryFunction
	MakeAddLiquidityPayload(lhsAmt, rhsAmt UITokenAmount) transactionbuilder.TransactionPayloadEntryFunction
	MakeRemoveLiquidityPayload(liqiudityAmt, lhsMinAmt, rhsMinAmt UITokenAmount) transactionbuilder.TransactionPayloadEntryFunction
}

type HippoPool struct {
	HippoPoolInterface
	TradeRoute
	LpCoinInfo types.CoinInfo
}

func NewHippoPool(xCoinInfo, yCoinInfo, LpCoinInfo types.CoinInfo) HippoPool {
	return HippoPool{
		LpCoinInfo: LpCoinInfo,
		TradeRoute: NewTradeRoute(xCoinInfo, yCoinInfo),
	}
}

func (h *HippoPool) LpTag() string {
	return "" // todo return LpCoinInfo.TokenType
}

func (h *HippoPool) XYFullName() string {
	return "" // todo
}

func (h *HippoPool) GetCurrentPrice() PriceType {
	return h.GetCurrentPriceDirectional(true)
}
func (h *HippoPool) GetQuote(uiAmount UITokenAmount) QuoteType {
	return h.GetQuoteDirectional(uiAmount, true)
}

func (h *HippoPool) MakeSwapPayload(amountIn, minAmountOut UITokenAmount) transactionbuilder.TransactionPayloadEntryFunction {
	return h.MakeSwapPayloadDirectional(amountIn, minAmountOut, true)
}

type RouteStep struct {
	Pool   HippoPool
	IsXtoY bool
}

func (r RouteStep) LhsTokenInfo() types.CoinInfo {
	if r.IsXtoY {
		return r.Pool.XCoinInfo
	} else {
		return r.Pool.YCoinInfo
	}
}

func (r RouteStep) RhsTokenInfo() types.CoinInfo {
	if r.IsXtoY {
		return r.Pool.YCoinInfo
	} else {
		return r.Pool.XCoinInfo
	}
}

type SteppedRoute struct {
	TradeRoute
	Steps []RouteStep
}

func NewSteppedRoute(steps []RouteStep) SteppedRoute {
	if len(steps) < 1 {
		panic("steps is nil")
	}
	firstStep := steps[0]
	lastStep := steps[len(steps)-1]
	return SteppedRoute{
		Steps:      steps,
		TradeRoute: NewTradeRoute(firstStep.LhsTokenInfo(), lastStep.RhsTokenInfo()),
	}
}

func (s *SteppedRoute) GetCurrentPrice() PriceType {
	var (
		xToY int64 = 1
		yToX int64 = 1
	)
	for _, step := range s.Steps {
		price := step.Pool.GetCurrentPriceDirectional(step.IsXtoY)
		xToY *= price.XToY
		yToX *= price.YToX
	}
	return PriceType{
		XToY: xToY,
		YToX: yToX,
	}
}

func (s *SteppedRoute) GetQuote(uiAmount UITokenAmount) QuoteType {
	if len(s.Steps) == 1 {
		return s.Steps[0].Pool.GetQuoteDirectional(uiAmount, s.Steps[0].IsXtoY)
	}
	prevOutputUiAmt := uiAmount
	avgPrice := int64(1)
	initialPrice := int64(1)
	finalPrice := int64(1)
	var quotes []QuoteType
	for _, step := range s.Steps {
		quote := step.Pool.GetQuoteDirectional(prevOutputUiAmt, step.IsXtoY)
		quotes = append(quotes, quote)
		prevOutputUiAmt = quote.OutputUiAmt
		avgPrice *= quote.AvgPrice
		initialPrice *= quote.InitialPrice
		finalPrice *= quote.FinalPrice
	}
	return QuoteType{
		InputSymbol:  s.XCoinInfo.Symbol,
		OutputSymbol: s.YCoinInfo.Symbol,
		AvgPrice:     avgPrice,
		InitialPrice: initialPrice,
		FinalPrice:   finalPrice,
		InputUiAmt:   uiAmount,
		OutputUiAmt:  prevOutputUiAmt,
		PriceImpact:  (finalPrice - initialPrice) / initialPrice,
	}
}

func (s *SteppedRoute) MakeSwapPayload(amountIn, minAmountOut UITokenAmount) transactionbuilder.TransactionPayloadEntryFunction {
	switch len(s.Steps) {
	case 1:
		return s.Steps[0].Pool.MakeSwapPayloadDirectional(amountIn, minAmountOut, s.Steps[0].IsXtoY)
	case 2:
		var (
			fromTokenInfo   = s.Steps[0].LhsTokenInfo()
			middleTokenInfo = s.Steps[0].RhsTokenInfo()
			toTokenInfo     = s.Steps[1].RhsTokenInfo()
			fromRawAmount   *big.Int
			toRawAmount     *big.Int
		)
		fromRawAmount = new(big.Int).Mul(amountIn, decimal.NewFromFloat(math.Pow(10, float64(fromTokenInfo.Decimals))).BigInt())
		toRawAmount = new(big.Int).Mul(minAmountOut, decimal.NewFromFloat(math.Pow(10, float64(toTokenInfo.Decimals))).BigInt())
		return transactionbuilder.TransactionPayloadEntryFunction{} // todo 组装结构体
	case 3:
		var (
			fromTokenInfo    = s.Steps[0].LhsTokenInfo()
			middle1TokenInfo = s.Steps[0].RhsTokenInfo()
			middle2TokenInfo = s.Steps[1].RhsTokenInfo()
			toTokenInfo      = s.Steps[2].RhsTokenInfo()
			fromRawAmount    *big.Int
			toRawAmount      *big.Int
		)
		fromRawAmount = new(big.Int).Mul(amountIn, decimal.NewFromFloat(math.Pow(10, float64(fromTokenInfo.Decimals))).BigInt())
		toRawAmount = new(big.Int).Mul(minAmountOut, decimal.NewFromFloat(math.Pow(10, float64(toTokenInfo.Decimals))).BigInt())
		return transactionbuilder.TransactionPayloadEntryFunction{} // todo 组装结构体
	default:
		panic("err")
	}
}

func (s *SteppedRoute) Concat(next SteppedRoute) SteppedRoute {
	if s.YCoinInfo.Symbol != next.XCoinInfo.Symbol {
		panic("Unable to join incompatible eroutes")
	}
	return NewSteppedRoute(append(s.Steps, next.Steps...))
}
func (s *SteppedRoute) GetAllPools() []HippoPool {
	var pools []HippoPool
	for _, step := range s.Steps {
		pools = append(pools, step.Pool)
	}
	return pools
}

func (s *SteppedRoute) GetSymbolPath() []string {
	symbols := []string{s.Steps[0].LhsTokenInfo().Symbol}
	for _, step := range s.Steps {
		if step.LhsTokenInfo().Symbol != symbols[len(symbols)-1] {
			panic("Bad path")
		}
		symbols = append(symbols, step.RhsTokenInfo().Symbol)
	}
	return symbols
}
func (s *SteppedRoute) Summarize() string {
	return strings.Join(s.GetSymbolPath(), " -> ")
}
