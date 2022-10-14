package pontem

import (
	"context"
	"testing"

	"github.com/coming-chat/go-aptos/aptosclient"
	"github.com/omnibtc/go-hippo-sdk/aggregator/coinlist"
	"github.com/omnibtc/go-hippo-sdk/contract"
)

const TestNode = "https://fullnode.testnet.aptoslabs.com"

func TestPoolProvider_LoadPoolList(t *testing.T) {
	client, err := aptosclient.Dial(context.Background(), TestNode)
	if err != nil {
		panic(err)
	}

	coinListClient, err := coinlist.LoadCoinListClient(contract.App{
		CoinList: contract.NewDevCoinListApp(),
	})
	if err != nil {
		panic(err)
	}
	type fields struct {
		client         *aptosclient.RestClient
		ownerAddress   string
		coinlistClient *coinlist.CoinListClient
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "test",
			fields: fields{
				client:         client,
				ownerAddress:   "0x385068db10693e06512ed54b1e6e8f1fb9945bb7a78c28a45585939ce953f99e",
				coinlistClient: coinListClient,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &PoolProvider{
				client:         tt.fields.client,
				ownerAddress:   tt.fields.ownerAddress,
				coinListClient: tt.fields.coinlistClient,
			}
			if got := p.LoadPoolList(); len(got) == 0 {
				t.Errorf("Pontem PoolProvider.LoadPoolList() Empty")
			}
		})
	}
}
