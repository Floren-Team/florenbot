package structs

import "time"

type Role struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	ShortName string    `gorm:"size:255;not null" json:"shortname"`
	BaseShort string    `gorm:"size:255;not null;column:base_short" json:"base_short"`
	ChatID    uint64    `json:"chat_id"`
	Chat      Chat      `gorm:"foreignKey:ChatID;references:ID" json:"chat"`
	CreatedAt time.Time `json:"created_at"`
}
