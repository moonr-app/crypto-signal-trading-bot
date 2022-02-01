package main

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gateio/gateapi-go/v6"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/moonr-app/crypto-signal-trading-bot/internal/exchange"
	"github.com/moonr-app/crypto-signal-trading-bot/internal/notifier"
	"github.com/moonr-app/crypto-signal-trading-bot/internal/persistence"
	"github.com/moonr-app/crypto-signal-trading-bot/internal/scraper"
	"github.com/moonr-app/crypto-signal-trading-bot/internal/shared/logging"
	"github.com/moonr-app/crypto-signal-trading-bot/internal/trader"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	botOwner := os.Getenv("BOT_OWNER")

	gateapiKey := os.Getenv("GATE_API_KEY")
	gateapiSecret := os.Getenv("GATE_API_SECRET")

	disableTelegram := os.Getenv("DISABLE_TELEGRAM")
	enableTestMode := os.Getenv("ENABLE_TEST_MODE")

	dynamoID := os.Getenv("DYNAMO_ID")
	dynamoSecret := os.Getenv("DYNAMO_SECRET")
	dynamoRegion := os.Getenv("DYNAMO_REGION")

	buyConsiderIntervalInSeconds := os.Getenv("BUY_INTERVAL_SECONDS")
	SellConsiderIntervalInSeconds := os.Getenv("SEll_INTERVAL_SECONDS")
	tickerCacheIntervalInSeconds := os.Getenv("TICKER_CACHE_INTERVAL_SECONDS")
	sellThresholdPercentage := os.Getenv("SELL_THRESHOLD_PERCENTAGE")
	toSpend := os.Getenv("USDT_TO_SPEND")

	spendableUSDT, err := decimal.NewFromString(toSpend)
	if err != nil {
		logging.Fatal(ctx, "failed to parse USDT_TO_SPEND", zap.String("value_passed", toSpend))
	}

	sellThreshAsFloat, err := strconv.ParseInt(sellThresholdPercentage, 10, 64)
	if err != nil {
		logging.Fatal(ctx, "failed to parse sellThresholdPercentage", zap.Error(err))
	}

	disableTeleBool, err := strconv.ParseBool(disableTelegram)
	if err != nil {
		logging.Fatal(ctx, "failed to parse failed to parse disableTelegram", zap.Error(err))
	}

	testmode, err := strconv.ParseBool(enableTestMode)
	if err != nil {
		logging.Fatal(ctx, "failed to parse failed to parse testmode", zap.Error(err))
	}

	buyConsiderInterval, err := strconv.ParseFloat(buyConsiderIntervalInSeconds, 10)
	if err != nil {
		logging.Fatal(ctx, "failed to parse failed to parse buyConsiderInterval", zap.Error(err))
	}

	sellConsiderInterval, err := strconv.ParseFloat(SellConsiderIntervalInSeconds, 10)
	if err != nil {
		logging.Fatal(ctx, "failed to parse failed to parse sellConsiderInterval", zap.Error(err))
	}

	tickerCacheInterval, err := strconv.ParseFloat(tickerCacheIntervalInSeconds, 10)
	if err != nil {
		logging.Fatal(ctx, "failed to parse failed to parse sellConsiderInterval", zap.Error(err))
	}

	var (
		buyConsiderIntervalSecs  = time.Duration(float64(time.Second) * buyConsiderInterval)
		sellConsiderIntervalSecs = time.Duration(float64(time.Second) * sellConsiderInterval)
		tickerCacheIntervalSecs  = time.Duration(float64(time.Second) * tickerCacheInterval)
		doer                     = http.DefaultClient
		db                       = persistence.NewDynamo(dynamoID, dynamoSecret, dynamoRegion)
		binance                  = scraper.NewBinance(doer)
		binanceCZ                = scraper.NewBinanceCZ(doer)
	)

	logging.Info(ctx, "running with threshold", zap.Int64("treshold", sellThreshAsFloat))

	coinbase, err := scraper.NewCoinbase(doer)
	if err != nil {
		logging.Fatal(ctx, "failed to init coinbase", zap.Error(err))
	}

	telegram := notifier.NewTelegram(doer, botOwner, disableTeleBool)

	ctx = context.WithValue(ctx, gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    gateapiKey,
		Secret: gateapiSecret,
	})

	gate, err := exchange.NewGateIO(ctx, testmode, spendableUSDT, tickerCacheIntervalSecs)
	if err != nil {
		logging.Fatal(ctx, "failed to create gate.io client", zap.Error(err))
	}

	var (
		buyer  = trader.NewBuyer(db, telegram, gate)
		seller = trader.NewSeller(telegram, db, gate, sellThreshAsFloat)
		t      = trader.NewTrader(buyConsiderIntervalSecs, sellConsiderIntervalSecs, buyer, seller, binance, coinbase, binanceCZ)
	)

	if err := t.Trade(ctx); err != nil {
		logging.Fatal(ctx, "unexpected trading error", zap.Error(err))
	}
}
