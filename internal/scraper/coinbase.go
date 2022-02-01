package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type coinbaseRes []struct {
	ID                    string `json:"id"`
	BaseCurrency          string `json:"base_currency"`
	QuoteCurrency         string `json:"quote_currency"`
	BaseMinSize           string `json:"base_min_size"`
	BaseMaxSize           string `json:"base_max_size"`
	QuoteIncrement        string `json:"quote_increment"`
	BaseIncrement         string `json:"base_increment"`
	DisplayName           string `json:"display_name"`
	MinMarketFunds        string `json:"min_market_funds"`
	MaxMarketFunds        string `json:"max_market_funds"`
	MarginEnabled         bool   `json:"margin_enabled"`
	FxStablecoin          bool   `json:"fx_stablecoin"`
	MaxSlippagePercentage string `json:"max_slippage_percentage"`
	PostOnly              bool   `json:"post_only"`
	LimitOnly             bool   `json:"limit_only"`
	CancelOnly            bool   `json:"cancel_only"`
	TradingDisabled       bool   `json:"trading_disabled"`
	Status                string `json:"status"`
	StatusMessage         string `json:"status_message"`
	AuctionMode           bool   `json:"auction_mode"`
}

const url = "https://api.exchange.coinbase.com/products"

type Coinbase struct {
	doer       Doer
	knownCoins map[string]struct{}
}

func (c *Coinbase) Name() string {
	return "coinbase"
}

func NewCoinbase(doer Doer) (*Coinbase, error) {
	m := make(map[string]struct{}, 0)

	c := &Coinbase{doer: doer, knownCoins: m}

	coins, err := c.getAllCoins(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to init coinbase  %w", err)
	}

	for _, coin := range coins {
		c.knownCoins[coin.BaseCurrency] = struct{}{}
	}
	return c, nil
}

func (c *Coinbase) getAllCoins(ctx context.Context) (coinbaseRes, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("coinbase: failed to create request: %w", err)
	}
	req.Header.Add("Accept", "application/json")

	res, err := c.doer.Do(req)
	if err != nil {
		return nil, fmt.Errorf("coinbase: failed to do request: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad response from coinbase: %d", res.StatusCode)
	}

	var cbr coinbaseRes
	if err := json.NewDecoder(res.Body).Decode(&cbr); err != nil {
		return nil, fmt.Errorf("failed to decode resposne: %w", err)
	}

	return cbr, nil
}

func (c *Coinbase) Scrape(ctx context.Context) (coin string, err error) {
	coins, err := c.getAllCoins(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get all coins: %w", err)
	}

	for _, coin := range coins {
		symbol := coin.BaseCurrency
		_, ok := c.knownCoins[symbol]
		if !ok {
			c.knownCoins[symbol] = struct{}{}
			return symbol, nil
		}
	}
	return "", ErrNoCoin
}
