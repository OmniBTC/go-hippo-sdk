package aggregator

import (
	"math/big"
	"sort"
	"sync"

	"github.com/omnibtc/go-hippo-sdk/aggregator/base"
	"github.com/omnibtc/go-hippo-sdk/contract"
	"github.com/omnibtc/go-hippo-sdk/types"
)

type TradeAggregator struct {
	app           contract.App
	fetcher       types.SimulationKeys
	poolProviders []base.TradingPoolProvider
	allPolls      []base.TradingPool
	xToAnyPools   map[string][]base.TradingPool
}

func NewTradeAggregator(
	app contract.App,
	fetcher types.SimulationKeys,
	poolProviders []base.TradingPoolProvider) *TradeAggregator {
	aggregator := &TradeAggregator{
		app:           app,
		fetcher:       fetcher,
		poolProviders: poolProviders,
		allPolls:      make([]base.TradingPool, 0),
		xToAnyPools:   make(map[string][]base.TradingPool),
	}
	aggregator.LoadAllPoolLists()
	return aggregator
}

func (a *TradeAggregator) LoadAllPoolLists() {
	allPools := make([]base.TradingPool, 0)
	wg := sync.WaitGroup{}
	l := sync.Mutex{}
	for _, p := range a.poolProviders {
		wg.Add(1)
		go func(p base.TradingPoolProvider) {
			defer wg.Done()
			pls := p.LoadPoolList()
			if pls == nil {
				return
			}
			l.Lock()
			defer l.Unlock()
			allPools = append(allPools, pls...)
		}(p)
	}
	wg.Wait()

	xToAnyPools := make(map[string][]base.TradingPool)
	for _, p := range allPools {
		fullName := p.XCoinInfo().TokenType.GetFullName()
		if _, ok := xToAnyPools[fullName]; !ok {
			xToAnyPools[fullName] = []base.TradingPool{p}
		} else {
			ps := xToAnyPools[fullName]
			ps = append(ps, p)
			xToAnyPools[fullName] = ps
		}
	}
	a.xToAnyPools = xToAnyPools
	a.allPolls = allPools
}

func (a *TradeAggregator) GetXtoYDirectSteps(x, y types.CoinInfo, requireRouteable bool) []base.TradeStep {
	xFullName := x.TokenType.GetFullName()
	yFullName := y.TokenType.GetFullName()
	if xFullName == yFullName {
		panic("cannot swap same token")
	}

	steps := make([]base.TradeStep, 0)
	if xToYCandidates, ok := a.xToAnyPools[xFullName]; ok {
		for _, pool := range xToYCandidates {
			if requireRouteable && !pool.IsRoutable() {
				continue
			}

			if pool.YCoinInfo().TokenType.GetFullName() == yFullName {
				steps = append(steps, base.NewTradeStep(pool, true))
			}
		}
	}
	if yToXCandidates, ok := a.xToAnyPools[yFullName]; ok {
		for _, pool := range yToXCandidates {
			if requireRouteable && !pool.IsRoutable() {
				continue
			}
			if pool.YCoinInfo().TokenType.GetFullName() == xFullName {
				steps = append(steps, base.NewTradeStep(pool, false))
			}
		}
	}

	return steps
}

func (a *TradeAggregator) GetOneStepRoutes(x, y types.CoinInfo) []base.TradeRoute {
	xFullName := x.TokenType.GetFullName()
	if xFullName == y.TokenType.GetFullName() {
		panic("cannot swap same token")
	}

	steps := a.GetXtoYDirectSteps(x, y, false)
	routes := make([]base.TradeRoute, 0)
	for _, step := range steps {
		routes = append(routes, base.NewTradeRoute([]base.TradeStep{step}))
	}
	return routes
}

func (a *TradeAggregator) GetTwoStepRoutes(x, y types.CoinInfo) ([]base.TradeRoute, error) {
	xFullName := x.TokenType.GetFullName()
	yFullName := y.TokenType.GetFullName()
	result := make([]base.TradeRoute, 0)
	fullList, err := a.app.CoinList.QueryFetchFullList()
	if err != nil {
		return nil, err
	}
	for _, k := range fullList {
		kFullName := k.TokenType.GetFullName()
		if kFullName == xFullName || kFullName == yFullName {
			continue
		}

		// x-to-k
		xTokSteps := a.GetXtoYDirectSteps(x, k, true)
		if len(xTokSteps) == 0 {
			continue
		}

		// k-to-y
		kToySteps := a.GetXtoYDirectSteps(k, y, true)
		if len(kToySteps) == 0 {
			continue
		}

		for _, xToK := range xTokSteps {
			for _, kToy := range kToySteps {
				result = append(result, base.NewTradeRoute(
					[]base.TradeStep{xToK, kToy},
				))
			}
		}
	}
	return result, nil
}

func (a *TradeAggregator) GetThreeStepRoutes(x, y types.CoinInfo) ([]base.TradeRoute, error) {
	xFullName := x.TokenType.GetFullName()
	yFullName := y.TokenType.GetFullName()
	result := make([]base.TradeRoute, 0)
	fullList, err := a.app.CoinList.QueryFetchFullList()
	if err != nil {
		return nil, err
	}
	for _, k := range fullList {
		kFullName := k.TokenType.GetFullName()
		if kFullName == xFullName || kFullName == yFullName {
			continue
		}

		// x-to-k 2steps
		xtoKRoutes, err := a.GetTwoStepRoutes(x, k)
		if err != nil || len(xtoKRoutes) == 0 {
			continue
		}
		kToYSteps := a.GetXtoYDirectSteps(k, y, true)
		if len(kToYSteps) == 0 {
			continue
		}
		for _, xToKRoute := range xtoKRoutes {
			for _, kToY := range kToYSteps {
				result = append(result, base.NewTradeRoute([]base.TradeStep{
					xToKRoute.Steps[0],
					xToKRoute.Steps[1],
					kToY,
				}))
			}
		}
	}
	return result, nil
}

func (a *TradeAggregator) GetAllRoutes(x, y types.CoinInfo, maxSteps int, allowRoundTrip bool) ([]base.TradeRoute, error) {
	allRoutes := make([]base.TradeRoute, 0)
	if maxSteps >= 1 {
		rs := a.GetOneStepRoutes(x, y)
		allRoutes = append(allRoutes, rs...)
	}
	if maxSteps >= 2 {
		if rs, err := a.GetTwoStepRoutes(x, y); err != nil {
			return nil, err
		} else {
			allRoutes = append(allRoutes, rs...)
		}
	}
	if maxSteps >= 3 {
		if rs, err := a.GetThreeStepRoutes(x, y); err != nil {
			return nil, err
		} else {
			allRoutes = append(allRoutes, rs...)
		}
	}
	if !allowRoundTrip {
		result := make([]base.TradeRoute, 0, len(allRoutes))
		for _, item := range allRoutes {
			if item.HasRoundTrip() {
				continue
			}
			result = append(result, item)
		}
		return result, nil
	}
	return allRoutes, nil
}

func (a *TradeAggregator) GetQuotes(inputAmount *big.Int, x, y types.CoinInfo, maxSteps int, reloadState bool, allowRoundTrip bool) ([]*base.RouteAndQuote, error) {
	routes, err := a.GetAllRoutes(x, y, maxSteps, allowRoundTrip)
	if err != nil {
		return nil, err
	}

	result := make([]*base.RouteAndQuote, len(routes))
	for i, route := range routes {
		result[i] = &base.RouteAndQuote{
			Route: route,
			Quote: route.GetQuote(inputAmount),
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return ((*big.Int)(result[i].Quote.OutputAmount)).Cmp(result[j].Quote.OutputAmount) >= 0
	})
	return result, nil
}

func (a *TradeAggregator) GetBestQuote(inputAmount *big.Int, x, y types.CoinInfo, maxSteps int, reloadState bool, allowRoundTrip bool) (*base.RouteAndQuote, error) {
	quotes, err := a.GetQuotes(inputAmount, x, y, maxSteps, reloadState, allowRoundTrip)
	if err != nil {
		return nil, err
	}
	if len(quotes) == 0 {
		return nil, nil
	}
	return quotes[0], nil
}
