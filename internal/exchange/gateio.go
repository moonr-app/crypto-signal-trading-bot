package exchange

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/antihax/optional"
	"github.com/gateio/gateapi-go/v6"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/moonr-app/crypto-signal-trading-bot/internal/shared/logging"
)

var (
	currencyTradingPairFmtString = "%s_USDT"

	accountType  = "spot"
	sideTypeBuy  = "buy"
	sideTypeSell = "sell"

	timeInForceGoodToClose = "gtc"

	nilReturnCurr = decimal.NewFromFloat(0)
)

type GateIO struct {
	api         *gateapi.APIClient
	testMode    bool
	toSpend     decimal.Decimal
	pricesCache map[string]decimal.Decimal
	lock        *sync.Mutex
}

func NewGateIO(ctx context.Context, testMode bool, toSpend decimal.Decimal, cacheInterval time.Duration) (*GateIO, error) {
	if toSpend.LessThanOrEqual(decimal.NewFromInt(0)) {
		return nil, errors.New("cannot have a 0 or less value for toSpend ")
	}
	cfg := gateapi.NewConfiguration()

	client := gateapi.NewAPIClient(cfg)

	g := &GateIO{
		api:         client,
		testMode:    testMode,
		toSpend:     toSpend,
		lock:        &sync.Mutex{},
		pricesCache: make(map[string]decimal.Decimal),
	}

	go func() {
		ticker := time.NewTicker(cacheInterval)
		for {
			select {
			case <-ticker.C:
				tickers, _, err := client.SpotApi.ListTickers(ctx, &gateapi.ListTickersOpts{})
				if err != nil {
					logging.Error(ctx, "failed to list tickers", zap.Error(err))
				}

				g.lock.Lock()
				for _, t := range tickers {
					if strings.Contains(t.CurrencyPair, "USDT") && t.LowestAsk != "" {
						i, err := decimal.NewFromString(t.LowestAsk)
						if err != nil {
							logging.Error(ctx, "invalid ticker price",
								zap.String("ticker", t.CurrencyPair),
								zap.String("lowest_ask", t.LowestAsk),
							)
							continue
						}
						g.pricesCache[t.CurrencyPair] = i
					}
				}
				g.lock.Unlock()
			}
		}
	}()

	return g, nil
}

func (g *GateIO) CheckSupport(ctx context.Context, coin string) (bool, error) {
	cur, _, err := g.api.SpotApi.GetCurrency(ctx, coin)
	if err != nil {
		return false, fmt.Errorf("failed to check support for currency: %w", err)
	}

	return !cur.TradeDisabled, nil
}

func (g *GateIO) PurchaseCoin(ctx context.Context, coin string, lastPrice decimal.Decimal) (pricePurchased decimal.Decimal, amountPurchased decimal.Decimal, err error) {
	var (
		currencyPair = fmt.Sprintf(currencyTradingPairFmtString, coin)
		volume       = g.toSpend.Div(lastPrice)
	)

	if g.testMode {
		logging.Info(ctx, "test mode, not really trading")
		return lastPrice, volume, nil
	}

	order, _, err := g.api.SpotApi.CreateOrder(ctx, gateapi.Order{
		CurrencyPair: currencyPair,
		Account:      accountType,
		Side:         sideTypeBuy,
		TimeInForce:  timeInForceGoodToClose,
		Price:        lastPrice.String(),
		Amount:       volume.String(),
	})

	if err != nil {
		return nilReturnCurr, nilReturnCurr, err
	}

	pricePurchasedAt, err := decimal.NewFromString(order.Price)
	if err != nil {
		return nilReturnCurr, nilReturnCurr, err
	}

	amtPurchased, err := decimal.NewFromString(order.Amount)
	if err != nil {
		return nilReturnCurr, nilReturnCurr, err
	}
	return pricePurchasedAt, amtPurchased, nil
}

func (g *GateIO) GetBalanceForCoin(ctx context.Context, coin string) (decimal.Decimal, error) {
	bals, resp, err := g.api.SpotApi.ListSpotAccounts(ctx, nil)
	if err != nil {
		return nilReturnCurr, fmt.Errorf("failed to get-account balance: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nilReturnCurr, fmt.Errorf("no error but bad response: %d", resp.StatusCode)
	}

	for _, v := range bals {
		if strings.ToLower(v.Currency) == strings.ToLower(coin) {
			d, err := decimal.NewFromString(v.Available)
			if err != nil {
				return nilReturnCurr, errors.New("couldn't convert balance to decimal")
			}
			return d, nil
		}
	}
	return nilReturnCurr, fmt.Errorf("didn't find coin %s in balances", coin)
}

func (g *GateIO) GetLastPrice(ctx context.Context, coin string) (decimal.Decimal, error) {
	currencyPair := fmt.Sprintf("%s_USDT", coin)

	g.lock.Lock()
	price, ok := g.pricesCache[currencyPair]
	g.lock.Unlock()
	if ok {
		return price, nil
	}

	cp, _, err := g.api.SpotApi.GetCurrencyPair(ctx, currencyPair)
	if err != nil {
		return decimal.NewFromInt(0), err
	}

	ticks, _, err := g.api.SpotApi.ListTickers(ctx, &gateapi.ListTickersOpts{CurrencyPair: optional.NewString(cp.Id)})
	if err != nil {
		return decimal.NewFromInt(0), err
	}

	if len(ticks) < 1 {
		return decimal.Zero, fmt.Errorf("expected at least one ticker, got none")
	}

	logging.Info(
		ctx,
		"got tickers",
		zap.String("last", ticks[0].Last),
		zap.String("lowest_ask", ticks[0].LowestAsk),
		zap.String("highest_bid", ticks[0].HighestBid),
	)

	i, _ := decimal.NewFromString(ticks[0].Last)
	return i, nil
}

func (g *GateIO) Sell(ctx context.Context, coin string, amount decimal.Decimal, lastPrice decimal.Decimal) (decimal.Decimal, error) {

	currencyPair := fmt.Sprintf("%s_USDT", coin)

	if g.testMode {
		logging.Info(ctx, "test mode, not really trading")
		return amount, nil
	}

	bal, err := g.GetBalanceForCoin(ctx, coin)
	if err != nil {
		return nilReturnCurr, fmt.Errorf("failed to get coin balance:%w", err)
	}

	logging.Info(ctx, "about to try and sell", zap.String("amount", amount.String()))

	_, _, err = g.api.SpotApi.CreateOrder(ctx, gateapi.Order{
		CurrencyPair: currencyPair,
		Account:      accountType,
		Side:         sideTypeSell,
		TimeInForce:  timeInForceGoodToClose,
		Price:        lastPrice.String(),
		Amount:       bal.String(),
	})
	if err != nil {
		return decimal.NewFromInt(0), err
	}

	logging.Info(ctx, "and sold!")
	return amount, nil
}
