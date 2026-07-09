package structs

import "time"

type Chat struct {
    ID        uint64    `gorm:"primaryKey" json:"id"`
    Name      string    `gorm:"size:255;not null" json:"name"`
    UserID    uint64    `gorm:"column:owner_id" json:"owner_id"`
    CreatedAt time.Time `json:"created_at"`
    Roles     []Role    `gorm:"foreignKey:ChatID"`

	User      User      `gorm:"foreignKey:UserID;references:ID;constraint:false" json:"-"`
}	