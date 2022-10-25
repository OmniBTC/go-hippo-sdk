package contract

import (
	"github.com/omnibtc/go-hippo-sdk/types"
)

const devModuleAddress = "0x498d8926f16eb9ca90cab1b3a26aa6f97a080b3fcbe6e83ae150b7243a00fb68"
const mainModuleAddress = "0xf22bede237a07e121b56d91a491eb7bcdfd1f5907926a9e58338f964a01b17fa"
const devModuleCoinName = "devnet_coins"
const mainModuleCoinName = "asset"

type CoinListApp interface {
	QueryFetchFullList() (list []types.CoinInfo, err error)
}

type CustomCoinListApp struct {
	coinList []types.CoinInfo
}

func NewCustomCoinListApp(coinList []types.CoinInfo) CoinListApp {
	return &CustomCoinListApp{
		coinList: coinList,
	}
}

func (c *CustomCoinListApp) QueryFetchFullList() ([]types.CoinInfo, error) {
	return c.coinList, nil
}

func (c *CustomCoinListApp) Clear() {
	c.coinList = make([]types.CoinInfo, 0)
}

func (c *CustomCoinListApp) Append(coinList []types.CoinInfo) {
	c.coinList = append(c.coinList, coinList...)
}

type DevCoinListApp struct {
}

func NewDevCoinListApp() CoinListApp {
	return &DevCoinListApp{}
}

func (c *DevCoinListApp) QueryFetchFullList() (list []types.CoinInfo, err error) {
	coinList := make([]types.CoinInfo, 0)
	data := [][3]string{
		// name, symbol, struct_name
		// all decimal is 8
		{"devBTC", "devBTC", "DevnetBTC"},
		{"devDAI", "devDAI", "DevnetDAI"},
		{"devUSDC", "devUSDC", "DevnetUSDC"},
		{"devUSDT", "devUSDT", "DevnetUSDT"},
	}
	mainData := [][3]string{
		{"zWETH", "zWETH", "WETH"},
		{"zUSDT", "zUSDT", "USDT"},
		{"zUSDC", "zUSDC", "USDC"},
	}
	for _, item := range data {
		coinList = append(coinList, types.CoinInfo{
			Name:     item[0],
			Decimals: 8,
			Symbol:   item[1],
			TokenType: &types.StructTag{
				Address: devModuleAddress,
				Module:  devModuleCoinName,
				Name:    item[2],
			},
		})
	}
	for _, item := range mainData {
		coinList = append(coinList, types.CoinInfo{
			Name:     item[0],
			Decimals: 8,
			Symbol:   item[1],
			TokenType: &types.StructTag{
				Address: mainModuleAddress,
				Module:  mainModuleCoinName,
				Name:    item[2],
			},
		})
	}

	coinList = append(coinList, types.CoinInfo{
		Name:     "XBTC",
		Decimals: 8,
		Symbol:   "XBTC",
		TokenType: &types.StructTag{
			Address: "0x3b0a7c06837e8fbcce41af0e629fdc1f087b06c06ff9e86f81910995288fd7fb",
			Module:  "xbtc",
			Name:    "XBTC",
		},
	}, types.CoinInfo{
		Name:     "USDA",
		Symbol:   "USDA",
		Decimals: 6,
		TokenType: &types.StructTag{
			Address: "0x1000000f373eb95323f8f73af0e324427ca579541e3b70c0df15c493c72171aa",
			Module:  "usda",
			Name:    "USDA",
		},
	}, types.CoinInfo{
		Name:     "APT",
		Symbol:   "APT",
		Decimals: 8,
		TokenType: &types.StructTag{
			Address: "0x1",
			Module:  "aptos_coin",
			Name:    "AptosCoin",
		},
	})

	return coinList, nil
}
