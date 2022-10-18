package coinlist

import (
	"github.com/omnibtc/go-hippo-sdk/contract"
	"github.com/omnibtc/go-hippo-sdk/types"
)

type CoinListClient struct {
	fullNameToCoinInfo map[string]types.CoinInfo
	coinList           []types.CoinInfo

	app contract.App
}

func LoadCoinListClient(app contract.App) (*CoinListClient, error) {
	c := &CoinListClient{
		app:                app,
		fullNameToCoinInfo: make(map[string]types.CoinInfo, 0),
		coinList:           make([]types.CoinInfo, 0),
	}
	err := c.buildCache()
	return c, err
}

func (c *CoinListClient) HasTokenType(tokenType types.TokenType) bool {
	_, ok := c.fullNameToCoinInfo[tokenType.GetFullName()]
	return ok
}

func (c *CoinListClient) GetCoinInfoList() []types.CoinInfo {
	return c.coinList
}

func (c *CoinListClient) GetCoinInfoByType(tokenType types.TokenType) (types.CoinInfo, bool) {
	v, o := c.fullNameToCoinInfo[tokenType.GetFullName()]
	return v, o
}

func (c *CoinListClient) GetCoinInfoByFullName(fullName string) (types.CoinInfo, bool) {
	v, o := c.fullNameToCoinInfo[fullName]
	return v, o
}

func (c *CoinListClient) buildCache() error {
	fullList, err := c.app.CoinList.QueryFetchFullList()
	if err != nil {
		return err
	}
	for _, tokenInfo := range fullList {
		c.fullNameToCoinInfo[tokenInfo.TokenType.GetFullName()] = tokenInfo
	}
	c.coinList = fullList
	return nil
}
