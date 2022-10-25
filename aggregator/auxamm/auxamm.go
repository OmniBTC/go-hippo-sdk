package auxamm

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/omnibtc/go-hippo-sdk/aggregator/base"
	"github.com/omnibtc/go-hippo-sdk/aggregator/coinlist"
	"github.com/omnibtc/go-hippo-sdk/types"
	"github.com/omnibtc/go-hippo-sdk/util"
)

type TradingPool struct {
	xCoinInfo    types.CoinInfo
	yCoinInfo    types.CoinInfo
	feeBps       int
	frozen       bool
	coinXReserve *big.Int
	coinYReserve *big.Int
	ownerAddress string
}

func NewTradingPool() base.TradingPool {
	return &TradingPool{}
}

func (t *TradingPool) DexType() base.DexType {
	return base.Aux
}

func (t *TradingPool) PoolType() base.PoolType {
	return 0
}

func (t *TradingPool) IsRoutable() bool {
	return !t.frozen
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
	panic("not implemented")
}

func (t *TradingPool) GetQuote(inputAmount base.TokenAmount, isXToY bool) base.QuoteType {
	if !t.IsStateLoaded() {
		panic("aux pool not loaded")
	}
	inputTokenInfo := t.xCoinInfo
	outputTokenInfo := t.yCoinInfo
	reserveInAmt := t.coinXReserve
	reserveOutAmt := t.coinYReserve
	if !isXToY {
		inputTokenInfo, outputTokenInfo = outputTokenInfo, inputTokenInfo
		reserveInAmt, reserveOutAmt = reserveOutAmt, reserveInAmt
	}

	coinOutAmt := util.GetCoinOutWithFees(inputAmount, reserveInAmt, reserveOutAmt, int64(t.feeBps), 10000)

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

func (t *TradingPool) MakePayload(input base.TokenAmount, minOut base.TokenAmount) types.EntryFunctionPayload {
	panic("not implemented")
}

type AuxPoolProvider struct {
	client         *aptosclient.RestClient
	ownerAddress   string
	coinListClient *coinlist.CoinListClient
}

func NewPoolProvider(client *aptosclient.RestClient, ownerAddress string, coinListClient *coinlist.CoinListClient) base.TradingPoolProvider {
	return &AuxPoolProvider{
		client:         client,
		ownerAddress:   ownerAddress,
		coinListClient: coinListClient,
	}
}

func (p *AuxPoolProvider) LoadPoolList() []base.TradingPool {
	poolList := make([]base.TradingPool, 0)
	resources, err := p.client.GetAccountResources(p.ownerAddress)
	if err != nil {
		return poolList
	}

	for _, resource := range resources {
		if !strings.Contains(resource.Type, "amm::Pool") {
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
		feeBps, _ := strconv.Atoi(resource.Data["fee_bps"].(string))
		frozen := resource.Data["frozen"].(bool)

		poolList = append(poolList, &TradingPool{
			xCoinInfo:    xCoinInfo,
			yCoinInfo:    yCoinInfo,
			ownerAddress: p.ownerAddress,
			coinXReserve: xint,
			coinYReserve: yint,
			feeBps:       feeBps,
			frozen:       frozen,
		})
	}

	return poolList
}
