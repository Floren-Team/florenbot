package structs

import (
	"time"
)

type UserPromo struct {
	// ID — первичный ключ, уникальный идентификатор записи
	ID uint `gorm:"primaryKey" json:"id"`

	// Code — сам промокод, должен быть уникальным и обязательным к заполнению
	Code string `gorm:"size:255;unique;not null" json:"code"`

	// OwnerID — ID пользователя, создавшего промокод.
	// Уникален, так как по вашей логике у одного пользователя может быть только один активный код
	OwnerID int64 `gorm:"not null;unique" json:"owner_id"`

	// Amount — сумма бонуса или награды
	Amount float64 `gorm:"not null" json:"amount"`

	// CreatedAt — время создания записи (GORM заполняет автоматически)
	CreatedAt time.Time `json:"created_at"`


	// ExpiresAt — время истечения активности промокода
	ExpiresAt time.Time `gorm:"null;column:expires_at" json:"expires_at"`

	// Owner — связь с моделью User.
	// Указывает GORM, что поле OwnerID ссылается на модель User
	Owner User `gorm:"foreignKey:OwnerID"`
}
