package models

import "time"

type Status string

const (
	BitStatusLowerBid  Status = "LowerBid"
	BitStatusOutBidded Status = "OutBidded"
)

type Bid struct {
	ID              int       `json:"id"`
	Date            time.Time `json:"date"`
	BitcircleAmount float32   `json:"bitcircle_amount,omitempty"`
	MoneyAmount     float32   `json:"money_amount"`
	Status          Status    `json:"status"`
	Winner          bool      `json:"winner"`
	AuctionID       int       `json:"auction_id"`
	WalletID        int       `json:"wallet_id,omitempty"`
}
