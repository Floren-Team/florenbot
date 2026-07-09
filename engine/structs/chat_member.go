package structs

import (
	"time"
)

type Member struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ChatID    uint64    `gorm:"index;not null" json:"chat_id"`
	UserID    uint64    `gorm:"index;not null" json:"user_id"`
	RoleID    uint64    `gorm:"index;not null" json:"role_id"`
	Role   *Role   `gorm:"foreignKey:RoleID;references:ID"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Member) TableName() string {
	return "chat_members"
}
