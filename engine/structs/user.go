package structs

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	// gorm:"primaryKey" — указывает, что это первичный ключ
	// ID типа uint64 соответствует вашим данным
	ID        uint64 `gorm:"primaryKey" json:"id"`
	Username  string `gorm:"size:255" json:"username"`
	FirstName string `json:"first_name"`

	// Используем float64 для финансовой точности (стандарт для GORM/Go)
	Balance    float64 `gorm:"default:1000" json:"balance"`
	FlorenCoin float64 `gorm:"default:300000" json:"floren_coin"`
	Euro       float64 `gorm:"default:17800" json:"euro"`

	Role      string `gorm:"size:32;default:'user'" json:"role"`
	PromoCode string `gorm:"size:32" json:"promocode"`

	// Указатель *int64 позволяет GORM корректно записывать NULL, если клана нет
	ClanID *int64 `json:"clan_id"`
	// GORM автоматически свяжет ClanID с таблицей Clan, если вы добавите поле:
	// Clan Clan `gorm:"foreignKey:ClanID"`

	NegativeReputation int `gorm:"default:0" json:"negative_reputation"`
	PositiveReputation int `gorm:"default:0" json:"positive_reputation"`
	Status             int `gorm:"default:0" json:"status"`
	Wins               int `gorm:"default:0" json:"wins"`
	Losses             int `gorm:"default:0" json:"losses"`

	// Поля времени, которые GORM заполняет автоматически
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"` // Для мягкого удаления
}
