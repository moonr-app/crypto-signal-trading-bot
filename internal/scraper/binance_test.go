package scraper_test

import (
	"bytes"
	"context"
	"errors"

	"io"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/moonr-app/crypto-signal-trading-bot/internal/mocks"
	"github.com/moonr-app/crypto-signal-trading-bot/internal/scraper"
	"github.com/stretchr/testify/require"
)

func TestBinance_Scrape(t *testing.T) {
	t.Run("returns an error given failure to scrape", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			doer = mocks.NewMockDoer(ctrl)
		)
		defer ctrl.Finish()

		binanceScraper := scraper.NewBinance(doer)

		doer.EXPECT().Do(gomock.Any()).Return(nil, errors.New("some-error"))

		coin, err := binanceScraper.Scrape(context.Background())
		require.Empty(t, coin)
		require.Error(t, err)
	})

	t.Run("returns error no coin given no match", func(t *testing.T) {
		var (
			ctrl = gomock.NewController(t)
			doer = mocks.NewMockDoer(ctrl)
		)
		defer ctrl.Finish()

		binanceScraper := scraper.NewBinance(doer)

		doer.EXPECT().Do(gomock.Any()).Return(&http.Response{
			Body: io.NopCloser(bytes.NewReader([]byte(`{"data":{"articles":[{"id":72200,"code":"e75ededcc356463a94786de743009a31","title":"Binance Adds SHIB/DOGE Trading Pair","body":null,"type":null,"catalogId":null,"catalogName":null,"publishDate":null}]}}`))),
		}, nil)

		coin, err := binanceScraper.Scrape(context.Background())
		require.Empty(t, coin)
		require.Error(t, err)
		require.True(t, errors.Is(err, scraper.ErrNoCoin))
	})

	t.Run("returns a match given varying title types", func(t *testing.T) {
		type testStruct struct {
			body         string
			expectedCoin string
		}
		tests := []testStruct{
			{
				body:         `{"data":{"articles":[{"id":72200,"code":"e75ededcc356463a94786de743009a31","title":"Binance Will List SuperRare (RARE)","body":null,"type":null,"catalogId":null,"catalogName":null,"publishDate":null}]}}`,
				expectedCoin: "rare",
			},
			{
				body:         `{"data":{"articles":[{"id":72200,"code":"e75ededcc356463a94786de743009a31","title":"Binance Will List Tranchess (CHESS)","body":null,"type":null,"catalogId":null,"catalogName":null,"publishDate":null}]}}`,
				expectedCoin: "chess",
			},
			{
				body:         `{"data":{"articles":[{"id":72200,"code":"e75ededcc356463a94786de743009a31","title":"Binance Will List Radicle (RAD)","body":null,"type":null,"catalogId":null,"catalogName":null,"publishDate":null}]}}`,
				expectedCoin: "rad",
			},
			{
				body:         `{"data":{"articles":[{"id":72200,"code":"e75ededcc356463a94786de743009a31","title":"Binance Will List Yield Guild Games (YGG)","body":null,"type":null,"catalogId":null,"catalogName":null,"publishDate":null}]}}`,
				expectedCoin: "ygg",
			},
			{
				body:         `{"data":{"articles":[{"id":72200,"code":"e75ededcc356463a94786de743009a31","title":"Binance Will List Ampleforth Governance Token (FORTH)","body":null,"type":null,"catalogId":null,"catalogName":null,"publishDate":null}]}}`,
				expectedCoin: "forth",
			},
		}

		var (
			ctrl = gomock.NewController(t)
			doer = mocks.NewMockDoer(ctrl)
		)
		defer ctrl.Finish()

		binanceScraper := scraper.NewBinance(doer)

		for _, v := range tests {
			doer.EXPECT().Do(gomock.Any()).Return(&http.Response{
				Body: io.NopCloser(bytes.NewReader([]byte(v.body))),
			}, nil)

			coin, err := binanceScraper.Scrape(context.Background())
			require.NotEmpty(t, coin)
			require.NoError(t, err)

			require.Equal(t, v.expectedCoin, coin)
		}

	})
}

func BenchmarkBinance_Scrape(b *testing.B) {
	binanceScraper := scraper.NewBinance(http.DefaultClient)
	ctx := context.Background()
	for i := 0; i < b.N; i++ {
		coin, err := binanceScraper.Scrape(ctx)
		if !errors.Is(err, scraper.ErrNoCoin) || coin != "" {
			b.Failed()
		}
	}
}
