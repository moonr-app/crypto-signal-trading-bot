package persistence

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/moonr-app/crypto-signal-trading-bot/internal/shared/logging"

	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/shopspring/decimal"

	"github.com/moonr-app/crypto-signal-trading-bot/internal/trader"
)

const (
	tableName          = "coin_history"
	statusAwaitingSale = "AWAITING_SALE"
	statusCompleted    = "COMPLETED"
	statusUnsupported  = "UNSUPPORTED"
)

type CoinItem struct {
	CoinSymbol     string
	PurchasePrice  string
	PurchaseAmount string
	PurchaseTime   time.Time
	TimeoutTime    time.Time
	PurchaseStatus string
}

type Dynamo struct {
	session *dynamodb.DynamoDB
}

func NewDynamo(id, secret, region string) *Dynamo {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentials(id, secret, ""),
			Region:      aws.String(region),
		},
	}))

	db := dynamodb.New(sess)
	return &Dynamo{session: db}
}

func (d *Dynamo) GetCoinsToConsider(ctx context.Context) ([]trader.SellingDetails, error) {
	filter := expression.Name("PurchaseStatus").Equal(expression.Value(statusAwaitingSale))
	expr, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		return nil, err
	}

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
	}

	result, err := d.session.ScanWithContext(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("query API call failed: %w", err)
	}

	var details []trader.SellingDetails
	for _, v := range result.Items {
		detail := CoinItem{}
		err = dynamodbattribute.UnmarshalMap(v, &detail)
		pamt, err := decimal.NewFromString(detail.PurchaseAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to convert purchaseAmount to decimal: %w", err)
		}
		pprice, err := decimal.NewFromString(detail.PurchasePrice)
		if err != nil {
			return nil, fmt.Errorf("failed to convert PurchasePrice to decimal: %w", err)
		}
		details = append(details, trader.SellingDetails{
			Coin:            detail.CoinSymbol,
			AmountPurchased: pamt,
			PurchaseTime:    detail.PurchaseTime,
			Timeout:         detail.TimeoutTime,
			PurchasePrice:   pprice,
		})
	}
	return details, nil
}

func (d *Dynamo) MarkCoinAsCompleted(ctx context.Context, coin string) error {
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":r": {
				S: aws.String(statusCompleted),
			},
		},
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"CoinSymbol": {
				S: aws.String(coin),
			},
		},
		ReturnValues:     aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String("set PurchaseStatus = :r"),
	}

	_, err := d.session.UpdateItem(input)
	if err != nil {
		logging.Fatal(ctx, "unexpected update item error", zap.Error(err))
	}
	return nil
}

func (d *Dynamo) CheckUniqueCoin(ctx context.Context, coin string) bool {
	filter := expression.Name("CoinSymbol").Equal(expression.Value(coin))
	expr, err := expression.NewBuilder().WithFilter(filter).Build()
	if err != nil {
		return false
	}

	params := &dynamodb.ScanInput{
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		ProjectionExpression:      expr.Projection(),
		TableName:                 aws.String(tableName),
	}

	result, err := d.session.ScanWithContext(ctx, params)
	if err != nil {
		return false
	}

	return len(result.Items) == 0
}

func (d Dynamo) StoreCoinUnsupported(ctx context.Context, coin string) error {
	c := CoinItem{
		CoinSymbol:     coin,
		PurchaseStatus: statusUnsupported,
	}
	av, err := dynamodbattribute.MarshalMap(c)
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = d.session.PutItemWithContext(ctx, input)
	if err != nil {
		return err
	}
	return nil
}

func (d *Dynamo) StoreCoinPurchased(ctx context.Context, coin string, purchasePrice decimal.Decimal, amountPurchased decimal.Decimal, timeout time.Time) error {
	c := CoinItem{
		CoinSymbol:     coin,
		PurchasePrice:  purchasePrice.String(),
		PurchaseAmount: amountPurchased.String(),
		PurchaseTime:   time.Now(),
		TimeoutTime:    timeout,
		PurchaseStatus: statusAwaitingSale,
	}
	av, err := dynamodbattribute.MarshalMap(c)
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = d.session.PutItemWithContext(ctx, input)
	if err != nil {
		return err
	}
	return nil
}
