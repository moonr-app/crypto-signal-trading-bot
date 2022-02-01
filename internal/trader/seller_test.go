package trader_test

import (
	"context"
	"errors"

	"testing"

	"github.com/golang/mock/gomock"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/moonr-app/crypto-signal-trading-bot/internal/mocks"
	"github.com/moonr-app/crypto-signal-trading-bot/internal/trader"
)

func TestSeller_MonitorAndSell(t *testing.T) {
	t.Run("err given cant call db for coins to consider", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			db = mocks.NewMockSellingDB(ctrl)
		)
		defer ctrl.Finish()

		s := trader.NewSeller(nil, db, nil, 0)

		db.EXPECT().GetCoinsToConsider(ctx).Return([]trader.SellingDetails{}, errors.New("err"))

		err := s.MonitorAndSell(ctx)
		require.Error(t, err)

		assert.Contains(t, err.Error(), "failed to read coins")
	})
	t.Run("err given cant get last price", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			db       = mocks.NewMockSellingDB(ctrl)
			exchange = mocks.NewMockSellingExchange(ctrl)

			coinToCheck = "mattcoin"
		)
		defer ctrl.Finish()

		s := trader.NewSeller(nil, db, exchange, 0)

		db.EXPECT().GetCoinsToConsider(ctx).Return([]trader.SellingDetails{{
			Coin: coinToCheck,
		}}, nil)
		exchange.EXPECT().GetLastPrice(ctx, coinToCheck).Return(decimal.NewFromFloat(0), errors.New("some-price-error"))

		err := s.MonitorAndSell(ctx)
		require.Error(t, err)

		assert.Contains(t, err.Error(), "failed to GetLastPrice")
	})
	t.Run("does not sell given less than threshold", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			db       = mocks.NewMockSellingDB(ctrl)
			exchange = mocks.NewMockSellingExchange(ctrl)

			coinToCheck              = "mattcoin"
			purchasePrice            = decimal.NewFromFloat(100)
			lastPrice                = decimal.NewFromFloat(150)
			purchaseThresholdPercent = int64(200)
		)
		defer ctrl.Finish()

		s := trader.NewSeller(nil, db, exchange, purchaseThresholdPercent)

		db.EXPECT().GetCoinsToConsider(ctx).Return([]trader.SellingDetails{{
			Coin:          coinToCheck,
			PurchasePrice: purchasePrice,
		}}, nil)
		exchange.EXPECT().GetLastPrice(ctx, coinToCheck).Return(lastPrice, nil)

		err := s.MonitorAndSell(ctx)
		require.NoError(t, err)
	})
	t.Run("sells given higher than threshold", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			ctx  = context.Background()

			db       = mocks.NewMockSellingDB(ctrl)
			exchange = mocks.NewMockSellingExchange(ctrl)
			notifier = mocks.NewMockNotifier(ctrl)

			coinToCheck = "mattcoin"

			amountToSell             = decimal.NewFromFloat(30)
			purchasePrice            = decimal.NewFromFloat(100)
			lastPrice                = decimal.NewFromFloat(300)
			purchaseThresholdPercent = int64(200)
		)
		defer ctrl.Finish()

		s := trader.NewSeller(notifier, db, exchange, purchaseThresholdPercent)

		gomock.InOrder(
			db.EXPECT().GetCoinsToConsider(ctx).Return([]trader.SellingDetails{{
				Coin:            coinToCheck,
				PurchasePrice:   purchasePrice,
				AmountPurchased: amountToSell,
			}}, nil),
			exchange.EXPECT().GetLastPrice(ctx, coinToCheck).Return(lastPrice, nil),
			exchange.EXPECT().Sell(ctx, coinToCheck, amountToSell, lastPrice).Return(amountToSell, nil),
			notifier.EXPECT().NotifySold(ctx, coinToCheck, amountToSell, lastPrice),
			db.EXPECT().MarkCoinAsCompleted(ctx, coinToCheck),
		)

		err := s.MonitorAndSell(ctx)
		require.NoError(t, err)
	})
	t.Run("notifies given timeout waiting to sell", func(t *testing.T) {})
}
