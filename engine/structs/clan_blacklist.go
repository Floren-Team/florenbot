package structs

import (
	"time"
)

// ClanBlacklist описывает таблицу черного списка клана
type ClanBlacklist struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ClanID    uint      `gorm:"not null;index" json:"clan_id"`
	UserID    int64     `gorm:"not null;index" json:"user_id"`
	Reason    string    `gorm:"size:255" json:"reason"`
	CreatedAt time.Time `json:"created_at"`

	// Связи для удобной подгрузки через Preload
	Clan Clan `gorm:"foreignKey:ClanID"`
	User User `gorm:"foreignKey:UserID"`
}
