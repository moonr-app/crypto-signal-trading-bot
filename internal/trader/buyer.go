package trader

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/moonr-app/crypto-signal-trading-bot/internal/shared/logging"

	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/moonr-app/crypto-signal-trading-bot/internal/scraper"
)

var ErrNoNewCoin = errors.New("coin is not new")
var ErrCoinUnsupported = errors.New("coin is not supported")

type Scraper interface {
	Scrape(ctx context.Context) (coin string, err error)
	Name() string
}

type PurchaseDB interface {
	CheckUniqueCoin(ctx context.Context, coin string) bool
	StoreCoinUnsupported(ctx context.Context, coin string) error
	StoreCoinPurchased(ctx context.Context, coin string, purchasePrice decimal.Decimal, amountPurchased decimal.Decimal, timeout time.Time) error
}

type ExchangePurchaser interface {
	CheckSupport(ctx context.Context, coin string) (bool, error)
	PurchaseCoin(ctx context.Context, coin string, lastPrice decimal.Decimal) (pricePurchased decimal.Decimal, amountPurchased decimal.Decimal, err error)
	GetLastPrice(ctx context.Context, coin string) (decimal.Decimal, error)
}
type Buyer struct {
	db              PurchaseDB
	notifier        Notifier
	exchange        ExchangePurchaser
	timeoutDuration time.Duration
}

func NewBuyer(db PurchaseDB, notifier Notifier, exchange ExchangePurchaser) *Buyer {
	return &Buyer{db: db, notifier: notifier, exchange: exchange}
}

func (b *Buyer) Buy(ctx context.Context, s Scraper) error {
	coin, err := s.Scrape(ctx)
	if err != nil {
		if errors.Is(err, scraper.ErrNoCoin) {
			return ErrNoNewCoin
		}

		return fmt.Errorf("error scraping: %w", err)
	}

	// see if we have a new coin.
	// if yes, check to see if we haven't seen it before.
	if newCoin := b.isCoinNew(ctx, coin); !newCoin {
		// if we have, log and do nothing.
		return ErrNoNewCoin
	}

	logging.Info(ctx, "new coin found", zap.String("name", s.Name()))

	// If we have not seen it before, check to see if we can purchase it on one of the supported exchanges.
	supported, err := b.exchange.CheckSupport(ctx, coin)
	if err != nil {
		return fmt.Errorf("failed to call exchange: %w", err)
	}

	// If not, log to say we couldn't; store coin in DB.
	if !supported {
		b.notifier.NotifyUnsupported(ctx, coin)
		if err := b.db.StoreCoinUnsupported(ctx, coin); err != nil {
			return fmt.Errorf("failed to store coin: %w", err)
		}
		return ErrCoinUnsupported
	}

	last, err := b.exchange.GetLastPrice(ctx, coin)
	if err != nil {
		logging.Error(ctx, "failed to get last price", zap.Error(err))
		return fmt.Errorf("failed to get last price: %w", err)
	}
	// if we can, make a purchase; store coin in DB.
	price, amount, err := b.exchange.PurchaseCoin(ctx, coin, last)
	if err != nil {
		return fmt.Errorf("failed to purchase coin: %w", err)
	}

	if err := b.db.StoreCoinPurchased(ctx, coin, price, amount, time.Now().Add(b.timeoutDuration)); err != nil {
		e := fmt.Errorf("failed to store coin purchase details: %w", err)
		return e
	}

	b.notifier.NotifyPurchased(ctx, coin, price, amount)

	return nil
}

func (b *Buyer) isCoinNew(ctx context.Context, coin string) bool {
	return b.db.CheckUniqueCoin(ctx, coin)
}
