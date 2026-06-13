package structs

type User struct {
    Id                 uint64  `json:"id"`
    Username           string  `json:"username"`
    Balance            float32 `json:"balance"`
    FlorenCoin         float32 `json:"floren_coin"`
    PromoCode          string  `json:"promocode"` 
    ClanId             *int64  `json:"clan_id"`   
    NegativeReputation int     `json:"negative_reputation"`
    PositiveReputation int     `json:"positive_reputation"`
    Status             int     `json:"status"`
}