package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"go.uber.org/zap"

	"github.com/moonr-app/crypto-signal-trading-bot/internal/shared/logging"
)

const (
	keyword    = "will list"
	matchRegex = "\\(([^)]+)"
)

var (
	r = regexp.MustCompile(matchRegex)
)

type binanceScrapeResponse struct {
	Data struct {
		Articles []struct {
			ID          int         `json:"id"`
			Code        string      `json:"code"`
			Title       string      `json:"title"`
			Body        interface{} `json:"body"`
			Type        interface{} `json:"type"`
			CatalogID   interface{} `json:"catalogId"`
			CatalogName interface{} `json:"catalogName"`
			PublishDate interface{} `json:"publishDate"`
		} `json:"articles"`
	} `json:"data"`
}

type Binance struct {
	currentPageSize int
	doer            Doer
}

func NewBinance(doer Doer) *Binance {
	return &Binance{doer: doer, currentPageSize: 1}
}

func (b *Binance) Name() string {
	return "binance"
}

func (b *Binance) Scrape(ctx context.Context) (coin string, err error) {
	if b.currentPageSize == 200 {
		b.currentPageSize = 1
	}
	url := fmt.Sprintf("https://www.binance.com/bapi/composite/v1/public/cms/article/catalog/list/query?catalogId=48&pageNo=1&pageSize=%d", b.currentPageSize)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create binance req: %w", err)
	}
	res, err := b.doer.Do(req)
	if err != nil {
		return "", fmt.Errorf("error doing: %w", err)
	}

	var scrapeRes binanceScrapeResponse
	if err := json.NewDecoder(res.Body).Decode(&scrapeRes); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	b.currentPageSize++

	if len(scrapeRes.Data.Articles) == 0 {
		return "", ErrNoCoin
	}

	lowerTitle := strings.ToLower(scrapeRes.Data.Articles[0].Title)

	if strings.Contains(lowerTitle, keyword) {
		logging.Info(ctx, "got a match!", zap.String("title", lowerTitle))
		s := r.FindString(lowerTitle)[1:]
		return s, nil
	}

	return "", ErrNoCoin
}
