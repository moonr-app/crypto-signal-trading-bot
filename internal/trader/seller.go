package trader

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/moonr-app/crypto-signal-trading-bot/internal/shared/logging"
)

type SellingDetails struct {
	Coin            string
	AmountPurchased decimal.Decimal
	PurchasePrice   decimal.Decimal
	PurchaseTime    time.Time
	Timeout         time.Time
}

type SellingDB interface {
	GetCoinsToConsider(ctx context.Context) ([]SellingDetails, error)
	MarkCoinAsCompleted(ctx context.Context, coin string) error
}

type SellingExchange interface {
	GetLastPrice(ctx context.Context, coin string) (decimal.Decimal, error)
	Sell(ctx context.Context, coin string, amount decimal.Decimal, lastPrice decimal.Decimal) (decimal.Decimal, error)
	GetBalanceForCoin(ctx context.Context, coin string) (decimal.Decimal, error)
}

type Seller struct {
	notifier                Notifier
	db                      SellingDB
	exchange                SellingExchange
	sellThresholdPercentage int64
}

func NewSeller(notifier Notifier, db SellingDB, exchange SellingExchange, sellThresholdPercentage int64) *Seller {
	return &Seller{notifier: notifier, db: db, exchange: exchange, sellThresholdPercentage: sellThresholdPercentage}
}

func (s *Seller) MonitorAndSell(ctx context.Context) error {
	coins, err := s.db.GetCoinsToConsider(ctx)
	if err != nil {
		return fmt.Errorf("failed to read coins from db: %w", err)
	}

	if len(coins) == 0 {
		logging.Debug(ctx, "no coins to consider")
	}

	for _, v := range coins {
		lastPrice, err := s.exchange.GetLastPrice(ctx, v.Coin)
		if err != nil {
			return fmt.Errorf("failed to GetLastPrice: %w", err)
		}

		logging.Info(
			ctx,
			"purchased coin",
			zap.String("current_price", v.PurchasePrice.String()),
			zap.String("last_price", lastPrice.String()),
		)

		if s.isGreaterThanSellThreshold(ctx, v.PurchasePrice, lastPrice) {
			sold, err := s.exchange.Sell(ctx, v.Coin, v.AmountPurchased, lastPrice)
			if err != nil {
				return fmt.Errorf("failed to sell coin: %w", err)
			}
			s.notifier.NotifySold(ctx, v.Coin, sold, lastPrice)
			if err := s.db.MarkCoinAsCompleted(ctx, v.Coin); err != nil {
				return fmt.Errorf("coin sold but couldn't mark it as so in DB: %w", err)
			}
		}
	}
	return nil
}

func (s *Seller) isGreaterThanSellThreshold(ctx context.Context, purchasePrice decimal.Decimal, lastPrice decimal.Decimal) bool {
	if purchasePrice.Equal(decimal.NewFromInt(0)) {
		logging.Warn(ctx, "purchase price was 0 for some reason")
		return false
	}

	var (
		percentIncrease = (lastPrice.Sub(purchasePrice)).Div(purchasePrice).Mul(decimal.NewFromInt(100))
		res             = percentIncrease.GreaterThanOrEqual(decimal.NewFromInt(s.sellThresholdPercentage))
	)

	logging.Info(
		ctx,
		"about to return, percentage increase",
		zap.String("percentage_increase", percentIncrease.String()),
		zap.Bool("result", res),
	)
	return res
}
