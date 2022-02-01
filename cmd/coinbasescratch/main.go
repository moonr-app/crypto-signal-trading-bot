package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/moonr-app/crypto-signal-trading-bot/internal/scraper"
	"github.com/moonr-app/crypto-signal-trading-bot/internal/shared/logging"
)

func main() {

	ctx := context.Background()
	c, err := scraper.NewCoinbase(http.DefaultClient)
	if err != nil {
		logging.Error(ctx, "could not create a new coinbase scraper", zap.Error(err))
		return
	}

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ticker.C:
			coin, err := c.Scrape(ctx)
			if err != nil {
				switch {
				case errors.Is(err, scraper.ErrNoCoin):
					logging.Info(ctx, "no new coin")
				default:
					logging.Error(ctx, "unexpected scraping error", zap.Error(err))
					return
				}
			}
			logging.Info(ctx, "new coin", zap.String("coin", coin))
		}
	}
}
