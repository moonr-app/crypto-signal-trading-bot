package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/moonr-app/crypto-signal-trading-bot/internal/shared/logging"
)

type binanceCZScrapeResponse struct {
	Data struct {
		Catalogs []struct {
			CatalogID       int         `json:"catalogId"`
			ParentCatalogID interface{} `json:"parentCatalogId"`
			Icon            string      `json:"icon"`
			CatalogName     string      `json:"catalogName"`
			Description     interface{} `json:"description"`
			CatalogType     int         `json:"catalogType"`
			Total           int         `json:"total"`
			Articles        []struct {
				ID          int    `json:"id"`
				Code        string `json:"code"`
				Title       string `json:"title"`
				Type        int    `json:"type"`
				ReleaseDate int64  `json:"releaseDate"`
			} `json:"articles"`
			Catalogs []interface{} `json:"catalogs"`
		} `json:"catalogs"`
	} `json:"data"`
}

type BinanceCZ struct {
	currentPageSize int
	doer            Doer
}

func NewBinanceCZ(doer Doer) *BinanceCZ {
	return &BinanceCZ{doer: doer, currentPageSize: 1}
}

func (b *BinanceCZ) Name() string {
	return "binanceCZ"
}

func (b *BinanceCZ) Scrape(ctx context.Context) (coin string, err error) {
	if b.currentPageSize == 200 {
		b.currentPageSize = 1
	}
	url := fmt.Sprintf("https://www.binancezh.com/gateway-api/v1/public/cms/article/list/query?catalogId=48&pageNo=1&type=1&pageSize=%d", b.currentPageSize)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create binance req: %w", err)
	}
	res, err := b.doer.Do(req)
	if err != nil {
		return "", fmt.Errorf("error doing: %w", err)
	}

	var scrapeRes binanceCZScrapeResponse
	if err := json.NewDecoder(res.Body).Decode(&scrapeRes); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	b.currentPageSize++

	for _, c := range scrapeRes.Data.Catalogs {
		if len(c.Articles) == 0 {
			continue
		}

		lowerTitle := strings.ToLower(c.Articles[0].Title)

		if strings.Contains(lowerTitle, keyword) {
			logging.Info(ctx, "got a match!", zap.String("title", lowerTitle))
			s := r.FindString(lowerTitle)[1:]
			return s, nil
		}
	}

	return "", ErrNoCoin
}
