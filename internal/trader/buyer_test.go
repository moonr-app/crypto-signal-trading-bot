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

func TestBuyer_Buy(t *testing.T) {
	t.Run("error given scraper fails", func(t *testing.T) {
		var (
			ctrl    = gomock.NewController(t)
			scraper = mocks.NewMockScraper(ctrl)
			ctx     = context.Background()
		)
		defer ctrl.Finish()

		b := trader.NewBuyer(nil, nil, nil)

		scraper.EXPECT().Scrape(ctx).Return("", errors.New("some-err"))

		err := b.Buy(ctx, scraper)
		require.Error(t, err)

		assert.Contains(t, err.Error(), "error scraping")
	})

	t.Run("ErrNoNewCoin given we get no new coin", func(t *testing.T) {
		var (
			ctrl    = gomock.NewController(t)
			scraper = mocks.NewMockScraper(ctrl)
			db      = mocks.NewMockPurchaseDB(ctrl)
			ctx     = context.Background()

			coinToCheck = "mattcoin"
		)
		defer ctrl.Finish()

		b := trader.NewBuyer(db, nil, nil)

		scraper.EXPECT().Scrape(ctx).Return(coinToCheck, nil)
		db.EXPECT().CheckUniqueCoin(ctx, coinToCheck).Return(false)

		err := b.Buy(ctx, scraper)
		require.Error(t, err)

		assert.True(t, errors.Is(err, trader.ErrNoNewCoin))
	})
	t.Run("err given we cant call exchange", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			scraper  = mocks.NewMockScraper(ctrl)
			db       = mocks.NewMockPurchaseDB(ctrl)
			notifier = mocks.NewMockNotifier(ctrl)
			exchange = mocks.NewMockExchangePurchaser(ctrl)

			ctx = context.Background()

			coinToCheck = "mattcoin"
		)
		defer ctrl.Finish()

		b := trader.NewBuyer(db, notifier, exchange)

		gomock.InOrder(
			scraper.EXPECT().Scrape(ctx).Return(coinToCheck, nil),
			db.EXPECT().CheckUniqueCoin(ctx, coinToCheck).Return(true),
			scraper.EXPECT().Name().Return("someScraper"),
			exchange.EXPECT().CheckSupport(ctx, coinToCheck).Return(false, errors.New("some-err")),
		)

		err := b.Buy(ctx, scraper)
		require.Error(t, err)

		assert.Contains(t, err.Error(), "failed to call exchange")
	})
	t.Run("NotifyUnsupported called given exchange doesnt support coin", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			scraper  = mocks.NewMockScraper(ctrl)
			db       = mocks.NewMockPurchaseDB(ctrl)
			notifier = mocks.NewMockNotifier(ctrl)
			exchange = mocks.NewMockExchangePurchaser(ctrl)

			ctx = context.Background()

			coinToCheck = "mattcoin"
		)
		defer ctrl.Finish()

		b := trader.NewBuyer(db, notifier, exchange)

		gomock.InOrder(
			scraper.EXPECT().Scrape(ctx).Return(coinToCheck, nil),
			db.EXPECT().CheckUniqueCoin(ctx, coinToCheck).Return(true),
			scraper.EXPECT().Name().Return("someScraper"),
			exchange.EXPECT().CheckSupport(ctx, coinToCheck).Return(false, nil),
			notifier.EXPECT().NotifyUnsupported(ctx, coinToCheck),
			db.EXPECT().StoreCoinUnsupported(ctx, coinToCheck).Return(nil),
		)

		err := b.Buy(ctx, scraper)
		require.Error(t, err)

		assert.True(t, errors.Is(err, trader.ErrCoinUnsupported))
	})
	t.Run("err given we cant purchase coin", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			scraper  = mocks.NewMockScraper(ctrl)
			db       = mocks.NewMockPurchaseDB(ctrl)
			notifier = mocks.NewMockNotifier(ctrl)
			exchange = mocks.NewMockExchangePurchaser(ctrl)

			ctx = context.Background()

			coinToCheck = "mattcoin"
			lastPrice   = decimal.NewFromFloat(32.3)
		)
		defer ctrl.Finish()

		b := trader.NewBuyer(db, notifier, exchange)

		gomock.InOrder(
			scraper.EXPECT().Scrape(ctx).Return(coinToCheck, nil),
			db.EXPECT().CheckUniqueCoin(ctx, coinToCheck).Return(true),
			scraper.EXPECT().Name().Return("someScraper"),
			exchange.EXPECT().CheckSupport(ctx, coinToCheck).Return(true, nil),
			exchange.EXPECT().GetLastPrice(ctx, coinToCheck).Return(lastPrice, nil),
			exchange.EXPECT().PurchaseCoin(ctx, coinToCheck, lastPrice).Return(decimal.NewFromFloat(0), decimal.NewFromFloat(0), errors.New("some-err")),
		)
		err := b.Buy(ctx, scraper)
		require.Error(t, err)

		assert.Contains(t, err.Error(), "failed to purchase coin")
	})
	t.Run("happy path; coin is purchased and notify is called", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			scraper  = mocks.NewMockScraper(ctrl)
			db       = mocks.NewMockPurchaseDB(ctrl)
			notifier = mocks.NewMockNotifier(ctrl)
			exchange = mocks.NewMockExchangePurchaser(ctrl)

			ctx = context.Background()

			coinToCheck     = "mattcoin"
			purchasePrice   = decimal.NewFromFloat(300)
			purchasedAmount = decimal.NewFromFloat(30)
			lastPrice       = decimal.NewFromFloat(69.69)
		)
		defer ctrl.Finish()

		b := trader.NewBuyer(db, notifier, exchange)

		gomock.InOrder(
			scraper.EXPECT().Scrape(ctx).Return(coinToCheck, nil),
			db.EXPECT().CheckUniqueCoin(ctx, coinToCheck).Return(true),
			scraper.EXPECT().Name().Return("myScraper"),
			exchange.EXPECT().CheckSupport(ctx, coinToCheck).Return(true, nil),
			exchange.EXPECT().GetLastPrice(ctx, coinToCheck).Return(lastPrice, nil),
			exchange.EXPECT().PurchaseCoin(ctx, coinToCheck, lastPrice).Return(purchasePrice, purchasedAmount, nil),
			db.EXPECT().StoreCoinPurchased(ctx, coinToCheck, purchasePrice, purchasedAmount, gomock.Any()).Return(nil),
			notifier.EXPECT().NotifyPurchased(ctx, coinToCheck, purchasePrice, purchasedAmount),
		)
		err := b.Buy(ctx, scraper)
		assert.NoError(t, err)
	})
}
