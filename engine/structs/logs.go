package structs

import "time"

type Log struct {
	ID uint `gorm:"primaryKey" json:"id"`
	ChatID uint64 `gorm:"index;not null" json:"chat_id"`
	UserID uint64 `gorm:"index;not null" json:"user_id"`
	User  User `gorm:"foreignKey:UserID"`
	Text string `gorm:"type:text;not null" json:"text"`
	CreatedAt time.Time `json:"created_at"`
}