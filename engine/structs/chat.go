package structs

import (
	"time"
)

type Chat struct {
	// ID — первичный ключ. Использование uint или int64 зависит от вашей базы,
	// но GORM по умолчанию ожидает ID.
	ID uint `gorm:"primaryKey" json:"id"`

	// Name — имя чата. Ограничим размер для оптимизации БД.
	Name string `gorm:"size:255;not null" json:"name"`

	// UserID — внешний ключ, ссылающийся на пользователя.
	// unique: true гарантирует, что у одного пользователя только один чат (как в вашем SQL).
	UserID int64 `gorm:"not null;uniqueIndex;column:owner_id" json:"owner_id"`

	// CreatedAt — время создания, заполняется автоматически.
	CreatedAt time.Time `json:"created_at"`

	// Связь с моделью User (позволяет делать .Preload("User"))
	User User `gorm:"foreignKey:UserID"`
}
