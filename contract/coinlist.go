package contract

import (
	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/omnibtc/go-hippo-sdk/types"
)

type CoinListApp struct {
	client *aptosclient.RestClient
}

func NewCoinListApp(client *aptosclient.RestClient) CoinListApp {
	return CoinListApp{
		client: client,
	}
}

func (c *CoinListApp) QueryFetchFullList() (list []types.CoinInfo, err error) {
	// payload := buildPayloadFetchFullList(listOwnerAddr)
	// var txs []*aptostypes.Transaction
	// txs, err = c.client.SimulateTransaction(&aptostypes.Transaction{
	// 	Payload: payload.ToAptosPayload(),
	// }, fetcher.Pubkey)
	// if err != nil {
	// 	return
	// }
	// todo
	return make([]types.CoinInfo, 0), nil
}

// func buildPayloadFetchFullList(listOwnerAddr string) types.EntryFunctionPayload {
// 	return types.EntryFunctionPayload{
// 		Function: fmt.Sprintf("%s::%s::%s", types.ModuleAddress, "coin_list", "fetch_full_list"),
// 		TypeArgs: []string{},
// 		Args:     []interface{}{listOwnerAddr},
// 	}
// }
