package structs

type UserCommandRestriction struct {
	ID       uint64 `gorm:"primaryKey"`
	UserID   uint64 `gorm:"index;not null"` // ID пользователя
	ChatID   uint64 `gorm:"index;not null"`
	Command string `gorm:"type:text;not null"`
}
