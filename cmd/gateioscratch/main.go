package main

import (
	"context"
	"os"
	"time"

	"github.com/shopspring/decimal"

	"github.com/gateio/gateapi-go/v6"
	"go.uber.org/zap"

	"github.com/moonr-app/crypto-signal-trading-bot/internal/exchange"
	"github.com/moonr-app/crypto-signal-trading-bot/internal/shared/logging"
)

func main() {
	const USDTCoin = "USDT"
	var spendable = decimal.NewFromInt(10)

	var (
		gateapiKey    = os.Getenv("GATE_API_KEY")
		gateapiSecret = os.Getenv("GATE_API_SECRET")
	)

	ctx := context.WithValue(context.Background(), gateapi.ContextGateAPIV4, gateapi.GateAPIV4{
		Key:    gateapiKey,
		Secret: gateapiSecret,
	})

	gate, err := exchange.NewGateIO(ctx, false, spendable, 500*time.Millisecond)
	if err != nil {
		logging.Fatal(ctx, "failed to create gate.io client", zap.Error(err))
	}

	bal, err := gate.GetBalanceForCoin(ctx, USDTCoin)
	if err != nil {
		logging.Error(ctx, "unable to get balance for coin", zap.String("coin", USDTCoin), zap.Error(err))
		return
	}

	logging.Info(ctx, "balance for coin", zap.String("balance", bal.String()))

	balLower, err := gate.GetBalanceForCoin(ctx, USDTCoin)
	if err != nil {
		logging.Error(ctx, "unable to get balance for coin", zap.String("coin", USDTCoin), zap.Error(err))
		return
	}

	logging.Info(ctx, "got lower balance", zap.String("balance", balLower.String()))
}
