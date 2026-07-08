package structs

import (
	"time"
)

type Clan struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"size:255;not null"`
	OwnerID     int64  `gorm:"not null"`
	MemberCount int    `gorm:"-"`
	OwnerName   string `gorm:"size:255;not null"`
	InviteCode  string `gorm:"size:32;unique"`
	CreatedAt   time.Time

	// Связи
	Members []User `gorm:"foreignKey:ClanID"`
}
