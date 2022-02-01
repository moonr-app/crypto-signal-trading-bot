package scraper

import (
	"errors"
	"net/http"
)

var (
	ErrNoCoin = errors.New("no new listing found")
)

type Doer interface {
	Do(req *http.Request) (*http.Response, error)
}
