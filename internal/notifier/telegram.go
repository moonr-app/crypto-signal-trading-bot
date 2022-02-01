package notifier

import (
	"context"
	"fmt"
	"net/http"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/moonr-app/crypto-signal-trading-bot/internal/shared/logging"
)

var (
	//TODO: add your telegram config here
	urlFmtString             = "https://api.telegram.org/$redacted/sendMessage?chat_id=-$redacted&text=%s&parse_mode=Markdown"
	errFmtString             = "[%s] An error occurred: %s"
	coinUnsupportedFmtString = "[%s] Wanted to buy coin %s but it was unsupported by gate.io :("
	purchaseFmtString        = "[%s] Just bought %s of %s at %s per coin."
	soldFmtString            = "[%s] Just sold %s of %s coin at %s per coin."
)

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}

type Telegram struct {
	url      string
	doer     Doer
	noOp     bool
	botOwner string
}

func NewTelegram(doer Doer, botOwner string, noOp bool) *Telegram {
	return &Telegram{
		url:      urlFmtString,
		doer:     doer,
		botOwner: botOwner,
		noOp:     noOp,
	}
}

func (t *Telegram) NotifyError(ctx context.Context, err error) {
	if t.noOp {
		return
	}
	text := fmt.Sprintf(errFmtString, t.botOwner, err.Error())
	urlWithText := fmt.Sprintf(urlFmtString, text)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlWithText, nil)
	if err != nil {
		logging.Error(ctx, "failed to build notify request", zap.Error(err))
		return
	}

	if _, err := t.doer.Do(req); err != nil {
		logging.Error(ctx, "failed to perform notify request", zap.Error(err))
	}
}

func (t *Telegram) NotifyUnsupported(ctx context.Context, coin string) {
	if t.noOp {
		return
	}
	text := fmt.Sprintf(coinUnsupportedFmtString, t.botOwner, coin)
	urlWithText := fmt.Sprintf(urlFmtString, text)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlWithText, nil)
	if err != nil {
		logging.Error(ctx, "failed to build notify unsupported request", zap.Error(err))
		return
	}

	if _, err := t.doer.Do(req); err != nil {
		logging.Error(ctx, "failed to perform notify unsupported request", zap.Error(err))
	}
}

func (t Telegram) NotifyPurchased(ctx context.Context, coin string, price decimal.Decimal, amount decimal.Decimal) {
	if t.noOp {
		return
	}
	text := fmt.Sprintf(purchaseFmtString, t.botOwner, amount, coin, price)
	urlWithText := fmt.Sprintf(urlFmtString, text)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlWithText, nil)
	if err != nil {
		logging.Error(ctx, "failed to build notify purchased request", zap.Error(err))
		return
	}

	res, err := t.doer.Do(req)
	if err != nil {
		logging.Error(ctx, "failed to perform notify purchased request", zap.Error(err))
		return
	}

	if res.StatusCode > 299 {
		logging.Warn(
			ctx,
			"got a bad response when performing notify purchased request",
			zap.Int("http_response_code", res.StatusCode),
			zap.Error(err),
		)
	}
}

func (t Telegram) NotifySold(ctx context.Context, coin string, amount decimal.Decimal, pricePerCoin decimal.Decimal) {
	if t.noOp {
		return
	}
	text := fmt.Sprintf(soldFmtString, t.botOwner, amount, coin, pricePerCoin)
	urlWithText := fmt.Sprintf(urlFmtString, text)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlWithText, nil)
	if err != nil {
		logging.Error(ctx, "failed to build notify sold request", zap.Error(err))
		return
	}

	if _, err := t.doer.Do(req); err != nil {
		logging.Error(ctx, "failed to perform notify sold request", zap.Error(err))
	}
}
