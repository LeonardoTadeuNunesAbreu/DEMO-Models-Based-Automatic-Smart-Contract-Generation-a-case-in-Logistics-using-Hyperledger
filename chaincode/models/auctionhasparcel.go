package models

type AuctionHasParcel struct {
	AuctionID int `json:"auction_id"`
	ParcelID  int `json:"parcel_id"`
}
