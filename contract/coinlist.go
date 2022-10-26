package contract

import (
	"encoding/json"
	"github.com/omnibtc/go-hippo-sdk/types"
	"io/ioutil"
	"os"
	"strings"
)

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
	coinFile, err := os.Open("contract/datajson.json")
	if err != nil {
		panic(err)
	}
	defer coinFile.Close()
	var arr []coinInfo
	coinValue, _ := ioutil.ReadAll(coinFile)
	err = json.Unmarshal(coinValue, &arr)
	if err != nil {
		return
	}

	for _, v := range arr {
		typesArr := strings.Split(v.TokenType.Type, "::")
		coinList = append(coinList, types.CoinInfo{
			Name:     v.Name,
			Decimals: int(v.Decimals),
			Symbol:   v.Symbol,
			TokenType: &types.StructTag{
				Address: typesArr[0],
				Module:  typesArr[1],
				Name:    typesArr[2],
			},
		})
	}

	return coinList, nil
}

type coinInfo struct {
	Name      string `json:"name"`
	Symbol    string `json:"symbol"`
	Decimals  int64  `json:"decimals"`
	LogoUrl   string `json:"logo_url"`
	TokenType struct {
		Type string `json:"type"`
	} `json:"token_type"`
}
