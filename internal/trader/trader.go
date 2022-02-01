package trader

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/moonr-app/crypto-signal-trading-bot/internal/shared/logging"
)

type Notifier interface {
	NotifyError(ctx context.Context, err error)
	NotifyUnsupported(ctx context.Context, coin string)
	NotifyPurchased(ctx context.Context, coin string, price decimal.Decimal, amount decimal.Decimal)
	NotifySold(ctx context.Context, coin string, amount decimal.Decimal, pricePerCoin decimal.Decimal)
}

type Trader struct {
	buyConsiderInterval  time.Duration
	sellConsiderInterval time.Duration
	buyer                *Buyer
	Seller               *Seller
	scrapers             []Scraper
}

func NewTrader(
	buyConsiderInterval time.Duration,
	sellConsiderInterval time.Duration,
	buyer *Buyer,
	seller *Seller,
	scrapers ...Scraper,
) *Trader {
	return &Trader{
		buyConsiderInterval:  buyConsiderInterval,
		sellConsiderInterval: sellConsiderInterval,
		buyer:                buyer,
		Seller:               seller,
		scrapers:             scrapers,
	}
}

func (t *Trader) Trade(ctx context.Context) error {
	buyTicker := time.NewTicker(t.buyConsiderInterval)
	sellTicker := time.NewTicker(t.sellConsiderInterval)

	g, ctx := errgroup.WithContext(ctx)

	for _, s := range t.scrapers {
		scraper := s

		g.Go(func() error {
			logger := logging.FromContext(ctx).With(zap.String("scraper", scraper.Name()))

			for {
				select {
				case <-buyTicker.C:
					err := t.buyer.Buy(ctx, scraper)
					if err != nil {
						switch {
						case errors.Is(err, ErrNoNewCoin):
							// do nothing
						default:
							logger.Error("buy error, should notify", zap.Error(err))
						}
					}
				case <-sellTicker.C:
					err := t.Seller.MonitorAndSell(ctx)
					if err != nil {
						logger.Error("sell error, should notify", zap.Error(err))
					}
				case <-ctx.Done():
					return errors.New("we are done")
				}
			}
		})
	}

	return g.Wait()
}
