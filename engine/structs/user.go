package structs

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	// ID типа uint64 соответствует формату Telegram ID
	ID        uint64 `gorm:"primaryKey;type:bigint unsigned" json:"id"`
	Username  string `gorm:"size:255" json:"username"`
	FirstName string `json:"first_name"`

	// Финансовые поля (float64 для простых расчетов)
	Balance    float64 `gorm:"default:1000" json:"balance"`
	FlorenCoin float64 `gorm:"default:300000" json:"floren_coin"`
	Euro       float64 `gorm:"default:17800" json:"euro"`

	RoleID *uint64 `gorm:"type:bigint unsigned" json:"role_id"`
	Role   *Role   `gorm:"foreignKey:RoleID;references:ID"`

	ClanID *uint64 `gorm:"type:bigint unsigned" json:"clan_id"`

	PromoCode string `gorm:"column:promocode;size:32" json:"promocode"`

	NegativeReputation int `gorm:"default:0" json:"negative_reputation"`
	PositiveReputation int `gorm:"default:0" json:"positive_reputation"`
	Status             int `gorm:"default:0" json:"status"`
	Wins               int `gorm:"default:0" json:"wins"`
	Losses             int `gorm:"default:0" json:"losses"`

	// Автоматические поля времени
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
