package structs

import "time"

type Ban struct {
	ID        uint64 `gorm:"primaryKey"`
	UserID    int64  `gorm:"index"`
	Reason    string
	CreatedAt time.Time
}
