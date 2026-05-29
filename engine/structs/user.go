package model

import (
	"database/sql"
)

type User struct {
	Id                 int     `json:"id"`
	Username           string  `json:"username"`
	Balance            float32 `json:"balance"`
	PromoCode          sql.NullString `json:"promocode"`
	ClanId             sql.NullInt64   `json:"clan_id"`
	Reputation         int     `json:"reputation"`
	NegativeReputation int     `json:"negative_reputation"`
	PositiveReputation int     `json:"positive_reputation"`
}
