package main

import (
	"context"
	"os"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/moonr-app/crypto-signal-trading-bot/internal/persistence"
	"github.com/moonr-app/crypto-signal-trading-bot/internal/shared/logging"
)

func main() {
	dynamoID := os.Getenv("DYNAMO_ID")
	dynamoSecret := os.Getenv("DYNAMO_SECRET")
	dynamoRegion := os.Getenv("DYNAMO_REGION")

	d := persistence.NewDynamo(dynamoID, dynamoSecret, dynamoRegion)
	ctx := context.Background()
	err := d.StoreCoinPurchased(ctx, "MAT2", decimal.NewFromInt(300), decimal.NewFromInt(200), time.Now().Add(time.Hour))
	if err != nil {
		logging.Fatal(ctx, "could not store purchased coin", zap.String("coin", "MAT2"), zap.Error(err))
	}

	err = d.StoreCoinPurchased(ctx, "MAT", decimal.NewFromInt(300), decimal.NewFromInt(200), time.Now().Add(time.Hour))
	if err != nil {
		logging.Fatal(ctx, "could not store purchased coin", zap.String("coin", "MAT"), zap.Error(err))
	}

	res, err := d.GetCoinsToConsider(ctx)
	if err != nil {
		logging.Fatal(ctx, "unexpected get coins to consider error", zap.Error(err))
	}

	logging.Info(ctx, "got coins to consider", zap.Any("coins", res))

	err = d.MarkCoinAsCompleted(ctx, "MAT2")
	if err != nil {
		logging.Fatal(ctx, "unexpected mark coin as completed error", zap.Error(err))
	}

	logging.Info(
		ctx,
		"check unique coin",
		zap.String("coin", "MAT2"),
		zap.Bool("is_unique", d.CheckUniqueCoin(ctx, "MAT2")),
	)
	logging.Info(
		ctx,
		"check unique coin",
		zap.String("coin", "MAT"),
		zap.Bool("is_unique", d.CheckUniqueCoin(ctx, "MAT")),
	)
	logging.Info(
		ctx,
		"check unique coin",
		zap.String("coin", "MAT"),
		zap.Bool("is_unique", d.CheckUniqueCoin(ctx, "MAT")),
	)

	err = d.StoreCoinUnsupported(ctx, "LMAO")
	if err != nil {
		logging.Fatal(ctx, "could not store coin", zap.String("coin", "LMAO"), zap.Error(err))
	}

	res, err = d.GetCoinsToConsider(ctx)
	if err != nil {
		logging.Fatal(ctx, "could not get coins to consider", zap.Error(err))
	}

	logging.Info(ctx, "result", zap.Any("res", res))
}
