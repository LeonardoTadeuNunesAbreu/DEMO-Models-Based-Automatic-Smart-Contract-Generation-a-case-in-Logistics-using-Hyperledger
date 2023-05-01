package models

import "time"

type Auction struct {
	ID            int       `json:"id"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	MaximumAmount float32   `json:"maximum_amount,omitempty"`
}
