package aptosswap

import (
	"errors"
	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/coming-chat/go-aptos/aptostypes"
	"github.com/omnibtc/go-hippo-sdk/aggregator/base"
	"github.com/omnibtc/go-hippo-sdk/aggregator/coinlist"
	"github.com/omnibtc/go-hippo-sdk/types"
	"math/big"
	"strings"
)

var (
	ZERO = big.NewInt(0)
)

type AptoswapSwapType string     //'v2' | 'stable'
type AptoswapFeeDirection string // 'X' | 'Y'

type AptoswapCoinType struct {
	Network string
	Name    string
}

type AptoswapPoolType struct {
	XTokenType AptoswapCoinType
	YTokenType AptoswapCoinType
}

type AptoswapPoolInfo struct {
	BpsScaling   *big.Int
	Type         AptoswapPoolType
	TypeString   string
	Index        *big.Int
	SwapType     AptoswapSwapType
	X            *big.Int
	Y            *big.Int
	LspSupply    *big.Int
	FeeDirection AptoswapFeeDirection
	Freeze       bool
	AdminFee     *big.Int
	LpFee        *big.Int
	IncentiveFee *big.Int
	ConnectFee   *big.Int
	WithdrawFee  *big.Int
}

func NewAptoswapPoolInfo(poolType AptoswapPoolType, typeString string, swapType AptoswapSwapType, feeDirection AptoswapFeeDirection, freeze bool, index, x, y, lspSupply, adminFee, lpFee, incentiveFee, connectFee, withdrawFee *big.Int) *AptoswapPoolInfo {
	pool := &AptoswapPoolInfo{}
	pool.Type = poolType
	pool.TypeString = typeString
	pool.Index = index
	pool.SwapType = swapType
	pool.X = x
	pool.Y = y
	pool.LspSupply = lspSupply
	pool.FeeDirection = feeDirection
	pool.Freeze = freeze
	pool.AdminFee = adminFee
	pool.LpFee = lpFee
	pool.IncentiveFee = incentiveFee
	pool.ConnectFee = connectFee
	pool.WithdrawFee = withdrawFee
	pool.BpsScaling = big.NewInt(10000)
	return pool
}

func MapResourceToPoolInfo(resource aptostypes.AccountResource) (*AptoswapPoolInfo, error) {
	var swapType AptoswapSwapType
	var feeDirection AptoswapFeeDirection

	typeString := resource.Type
	tag, err := types.ParseMoveStructTag(resource.Type)
	if err != nil {
		return nil, err
	}
	xCoinType := AptoswapCoinType{
		Network: "aptos",
		Name:    tag.TypeParams[0].StructTag.Name,
	}
	yCoinType := AptoswapCoinType{
		Network: "aptos",
		Name:    tag.TypeParams[1].StructTag.Name,
	}
	data := resource.Data
	poolType := AptoswapPoolType{
		XTokenType: xCoinType,
		YTokenType: yCoinType,
	}
	if data["pool_type"].(float64) == 100 {
		swapType = "v2"
	} else {
		swapType = "stable"
	}
	if data["fee_direction"].(float64) == 200 {
		feeDirection = "X"
	} else {
		feeDirection = "Y"
	}
	index, b := big.NewInt(0).SetString(data["index"].(string), 10)
	if !b {
		return nil, nil
	}
	x, b := big.NewInt(0).SetString(data["x"].(map[string]interface{})["value"].(string), 10)
	if !b {
		return nil, nil
	}
	y, b := big.NewInt(0).SetString(data["y"].(map[string]interface{})["value"].(string), 10)
	if !b {
		return nil, nil
	}
	if x.Cmp(big.NewInt(0)) == 0 || y.Cmp(big.NewInt(0)) == 0 {
		return nil, nil
	}
	lspSupply, b := big.NewInt(0).SetString(data["lsp_supply"].(string), 10)
	if !b {
		return nil, nil
	}
	freeze := data["freeze"].(bool)
	adminFee, b := big.NewInt(0).SetString(data["admin_fee"].(string), 10)
	if !b {
		return nil, nil
	}
	lpFee, b := big.NewInt(0).SetString(data["lp_fee"].(string), 10)
	if !b {
		return nil, nil
	}
	incentiveFee, b := big.NewInt(0).SetString(data["incentive_fee"].(string), 10)
	if !b {
		return nil, nil
	}
	connectFee, b := big.NewInt(0).SetString(data["connect_fee"].(string), 10)
	if !b {
		return nil, nil
	}
	withdrawFee, b := big.NewInt(0).SetString(data["withdraw_fee"].(string), 10)
	if !b {
		return nil, nil
	}
	return NewAptoswapPoolInfo(poolType, typeString, swapType, feeDirection, freeze, index, x, y, lspSupply, adminFee, lpFee, incentiveFee, connectFee, withdrawFee), nil
}

func (a *AptoswapPoolInfo) GetXToYAmount(dx *big.Int) *big.Int {
	xReserveAmt := a.X
	yReserveAmt := a.Y
	if a.FeeDirection == "X" {
		dx = new(big.Int).Sub(dx, new(big.Int).Div(new(big.Int).Mul(dx, a.TotalAdminFee()), a.BpsScaling))
	}
	dx = new(big.Int).Sub(dx, new(big.Int).Div(new(big.Int).Mul(dx, a.TotalLpFee()), a.BpsScaling))
	if dx.Cmp(ZERO) < 0 {
		return ZERO
	}
	dy := a._computeAmount(dx, xReserveAmt, yReserveAmt)
	if a.FeeDirection == "Y" {
		dy = new(big.Int).Sub(dy, new(big.Int).Div(new(big.Int).Mul(dy, a.TotalAdminFee()), a.BpsScaling))
	}
	return dy
}

func (a *AptoswapPoolInfo) GetYToXAmount(dy *big.Int) *big.Int {
	xReserveAmt := a.X
	yReserveAmt := a.Y
	if a.FeeDirection == "Y" {
		dy = new(big.Int).Sub(dy, new(big.Int).Div(new(big.Int).Mul(dy, a.TotalAdminFee()), a.BpsScaling))
	}
	dy = new(big.Int).Sub(dy, new(big.Int).Div(new(big.Int).Mul(dy, a.TotalLpFee()), a.BpsScaling))
	if dy.Cmp(ZERO) < 0 {
		return ZERO
	}
	dx := a._computeAmount(dy, yReserveAmt, xReserveAmt)
	if a.FeeDirection == "X" {
		dx = new(big.Int).Sub(dx, new(big.Int).Div(new(big.Int).Mul(dx, a.TotalAdminFee()), a.BpsScaling))
	}
	return dx
}

func (a *AptoswapPoolInfo) TotalAdminFee() *big.Int {
	return new(big.Int).Add(a.AdminFee, a.ConnectFee)
}

func (a *AptoswapPoolInfo) TotalLpFee() *big.Int {
	return new(big.Int).Add(a.IncentiveFee, a.LpFee)
}

func (a *AptoswapPoolInfo) _computeAmount(dx, x, y *big.Int) *big.Int {
	numerator := new(big.Int).Mul(y, dx)
	denominator := new(big.Int).Add(x, dx)
	dy := new(big.Int).Div(numerator, denominator)
	return dy
}

type AptoswapTradingPool struct {
	PackageAddr string
	_xCoinInfo  types.CoinInfo
	_yCoinInfo  types.CoinInfo
	Tag         types.StructTag
	Pool        *AptoswapPoolInfo
}

func NewAptoswapTradingPool(packageAddr string, _xCoinInfo, _yCoinInfo types.CoinInfo, tag types.StructTag, resource aptostypes.AccountResource) (*AptoswapTradingPool, error) {
	aptoswapTradingPool := &AptoswapTradingPool{
		PackageAddr: packageAddr,
		_xCoinInfo:  _xCoinInfo,
		_yCoinInfo:  _yCoinInfo,
		Tag:         tag,
	}
	pool, err := MapResourceToPoolInfo(resource)
	if err != nil {
		return nil, err
	}
	if pool == nil {
		return nil, errors.New("not found pool")
	}
	aptoswapTradingPool.Pool = pool
	return aptoswapTradingPool, nil
}

func (a *AptoswapTradingPool) DexType() base.DexType {
	return base.Aptosswap
}

func (a *AptoswapTradingPool) PoolType() base.PoolType {
	return 0
}

func (a *AptoswapTradingPool) IsRoutable() bool {
	return true
}

func (a *AptoswapTradingPool) XCoinInfo() types.CoinInfo {
	return a._xCoinInfo
}

func (a *AptoswapTradingPool) YCoinInfo() types.CoinInfo {
	return a._yCoinInfo
}

func (a *AptoswapTradingPool) IsStateLoaded() bool {
	return true
}

func (a *AptoswapTradingPool) GetPrice() base.PriceType {
	panic("not implemented")
}

func (a *AptoswapTradingPool) GetQuote(inputAmount base.TokenAmount, isXToY bool) base.QuoteType {
	if !a.IsStateLoaded() {
		panic("aptosswap pool not loaded")
	}
	inputTokenInfo := a._xCoinInfo
	outputTokenInfo := a._yCoinInfo
	if !isXToY {
		inputTokenInfo, outputTokenInfo = outputTokenInfo, inputTokenInfo
	}
	coinAmt := inputAmount
	coinOutAmt := a.Pool.GetXToYAmount(coinAmt)
	if !isXToY {
		coinOutAmt = a.Pool.GetYToXAmount(coinAmt)
	}
	outputUiAmt := coinOutAmt

	return base.QuoteType{
		InputSymbol:  inputTokenInfo.Symbol,
		OutputSymbol: outputTokenInfo.Symbol,
		InputAmount:  inputAmount,
		OutputAmount: outputUiAmt,
	}
}

func (a *AptoswapTradingPool) GetTagE() types.TokenType {
	return types.U8
}

func (a *AptoswapTradingPool) MakePayload(base.TokenAmount, base.TokenAmount) types.EntryFunctionPayload {
	panic("not implemented")
}

type AptoswapPoolProvider struct {
	client         *aptosclient.RestClient
	ownerAddress   string
	coinListClient *coinlist.CoinListClient
}

func NewPoolProvider(client *aptosclient.RestClient, ownerAddress string, coinListClient *coinlist.CoinListClient) base.TradingPoolProvider {
	return &AptoswapPoolProvider{
		client:         client,
		ownerAddress:   ownerAddress,
		coinListClient: coinListClient,
	}
}

func (p *AptoswapPoolProvider) SetResourceTypes(resourceTypes []string) {}

func (p *AptoswapPoolProvider) LoadPoolList() []base.TradingPool {
	poolList := make([]base.TradingPool, 0)
	resources, err := p.client.GetAccountResources(p.ownerAddress, 0)
	if err != nil {
		return poolList
	}

	for _, resource := range resources {
		if !strings.Contains(resource.Type, "pool::Pool") {
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
		pool, err := NewAptoswapTradingPool(p.ownerAddress, xCoinInfo, yCoinInfo, tag, resource)
		if err != nil {
			// todo handle error
			continue
		}

		poolList = append(poolList, pool)
	}

	return poolList
}
