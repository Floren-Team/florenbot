package structs

import (
	"database/sql"
)

type User struct {
	Id                 int            `json:"id"`
	Username           string         `json:"username"`
	Balance            float32        `json:"balance"`
	FlorenCoin         float32        `json:"floren_coin"`
	PromoCode          sql.NullString `json:"promocode"`
	ClanId             sql.NullInt64  `json:"clan_id"`
	NegativeReputation int            `json:"negative_reputation"`
	PositiveReputation int            `json:"positive_reputation"`
	Status             int            `json:"status"`
}
