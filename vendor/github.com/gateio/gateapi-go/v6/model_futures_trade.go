/*
 * Gate API v4
 *
 * Welcome to Gate.io API  APIv4 provides spot, margin and futures trading operations. There are public APIs to retrieve the real-time market statistics, and private APIs which needs authentication to trade on user's behalf.
 *
 * Contact: support@mail.gate.io
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package gateapi

type FuturesTrade struct {
	// Trade ID
	Id int64 `json:"id,omitempty"`
	// Trading time
	CreateTime float64 `json:"create_time,omitempty"`
	// Trading time, with milliseconds set to 3 decimal places.
	CreateTimeMs float64 `json:"create_time_ms,omitempty"`
	// Futures contract
	Contract string `json:"contract,omitempty"`
	// Trading size
	Size int64 `json:"size,omitempty"`
	// Trading price
	Price string `json:"price,omitempty"`
}