package docs

//go:generate mockgen -package mocks -destination internal/mocks/binance.go -source internal/scraper/scraper.go Doer
//go:generate mockgen -package mocks -destination internal/mocks/buyer.go  -source internal/trader/buyer.go Scraper,PurchaseDB,ExchangePurchaser
//go:generate mockgen -package mocks -destination internal/mocks/seller.go  -source internal/trader/seller.go SellingDB,SellingExchange
//go:generate mockgen -package mocks -destination internal/mocks/trader.go  -source internal/trader/trader.go Notifier
