/*
 * Gate API v4
 *
 * Welcome to Gate.io API  APIv4 provides spot, margin and futures trading operations. There are public APIs to retrieve the real-time market statistics, and private APIs which needs authentication to trade on user's behalf.
 *
 * Contact: support@mail.gate.io
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package gateapi

type SpotPriceTrigger struct {
	// Trigger price
	Price string `json:"price"`
	// Price trigger condition  - >=: triggered when market price larger than or equal to `price` field - <=: triggered when market price less than or equal to `price` field
	Rule string `json:"rule"`
	// How long (in seconds) to wait for the condition to be triggered before cancelling the order.
	Expiration int32 `json:"expiration"`
}
