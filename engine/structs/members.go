package structs

import (
	"time"
)

type ClanMember struct {
	// Составной первичный ключ
	ClanID uint  `gorm:"primaryKey" json:"clan_id"`
	UserID int64 `gorm:"primaryKey" json:"user_id"`

	// Дополнительные поля
	Role     string    `gorm:"size:20;default:'member'" json:"role"`
	JoinedAt time.Time `json:"joined_at"`

	// Явное описание связей для GORM
	Clan Clan `gorm:"foreignKey:ClanID"`
	User User `gorm:"foreignKey:UserID"`
}
