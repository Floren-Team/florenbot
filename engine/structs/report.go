package structs

import (
	"time"
)

type Report struct {
	ID uint `gorm:"primaryKey" json:"id"`

	// Используем указатели (*int64) вместо sql.NullInt64.
	// Если указатель nil, GORM запишет NULL в базу данных.
	// Это гораздо удобнее при работе с JSON и логикой в Go.
	AngryID *uint64 `json:"angry_id"`
	UserID  *uint64 `json:"user_id"`

	Text   string `gorm:"type:text;not null" json:"text"`
	Active bool   `gorm:"default:true" json:"active"`

	CreatedAt time.Time `json:"created_at"`

	// Связи (если нужно подгружать данные пользователя через .Preload)
	User User `gorm:"foreignKey:UserID"`
}
