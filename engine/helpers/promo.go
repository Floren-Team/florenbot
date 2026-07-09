package helpers

import (
	engine "florenbot/engine/mysql"
	"florenbot/engine/structs"
	"gorm.io/gorm"
	"log"
	"strings"
)

// ActivateCode обновляет промокод у пользователя
func ActivateCode(id uint64, code string) error {
	return engine.DB.Model(&structs.User{}).Where("id = ?", id).Update("promocode", code).Error
}

// CreateCode создает новый промокод
func CreateCode(code string, amount int, ownerID uint64) error {
	promo := structs.UserPromo{
		Code:    code,
		Amount:  float64(amount),
		OwnerID: int64(ownerID),
	}
	return engine.DB.Create(&promo).Error
}

func UpdatePromoExpire(code string, expiresAt string) error {
	return engine.DB.Model(&structs.UserPromo{}).Where("code = ?", code).Update("expires_at", expiresAt).Error
}

func DeletePromoExpire(code string) error {
	return engine.DB.Model(&structs.UserPromo{}).Where("code = ?", code).Update("expires_at", nil).Error
}

func GetPromocodesMemberCount() (map[string]int, error) {
	stats := make(map[string]int)

	// 1. Получаем список уникальных кодов
	var promoCodes []string
	err := engine.DB.Model(&structs.UserPromo{}).
		Distinct("code").
		Pluck("code", &promoCodes).Error

	if err != nil {
		return nil, err
	}

	// 2. Для каждого кода считаем количество пользователей
	for _, code := range promoCodes {
		cleanCode := strings.TrimSpace(code)
		if cleanCode == "" {
			continue
		}

		var count int64
		err := engine.DB.Model(&structs.User{}).
			Where("promocode = ?", cleanCode).
			Count(&count).Error

		if err != nil {
			log.Printf("Ошибка получения количества пользователей для промокода %s: %v", cleanCode, err)
			continue
		}

		// 3. Записываем в статистику
		stats[cleanCode] = int(count)
	}

	return stats, nil
}

// GetUserCode получает код пользователя
func GetUserCode(userID uint64) (string, error) {
	var user structs.User
	err := engine.DB.Select("promocode").First(&user, userID).Error
	return user.PromoCode, err
}

// DeleteCode удаляет промокод и очищает его у всех пользователей (в транзакции)
func DeleteCode(code string) error {
	return engine.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Удаляем сам промокод
		if err := tx.Where("code = ?", code).Delete(&structs.UserPromo{}).Error; err != nil {
			return err
		}
		// 2. Обнуляем у всех пользователей
		return tx.Model(&structs.User{}).Where("promocode = ?", code).Update("promocode", nil).Error
	})
}

// GetPromocodesUser возвращает данные промокода пользователя
func GetPromocodesUser(id uint64) ([]structs.UserPromo, error) {
	code, err := GetUserCode(id)
	if err != nil || code == "" {
		return nil, err
	}

	promo, err := GetCode(code)
	if err != nil {
		return nil, err
	}

	return []structs.UserPromo{*promo}, nil
}

// UpdatePromo обновляет данные промокода (в транзакции)
func UpdatePromo(code string, amount float64, userID uint64) error {
	return engine.DB.Transaction(func(tx *gorm.DB) error {
		// Обновляем промокод
		if err := tx.Model(&structs.UserPromo{}).Where("owner_id = ?", userID).
			Updates(map[string]interface{}{"code": code, "amount": int(amount)}).Error; err != nil {
			return err
		}
		// Обновляем у пользователя
		return tx.Model(&structs.User{}).Where("id = ?", userID).Update("promocode", code).Error
	})
}

// GetCode возвращает структуру промокода по его строковому значению
func GetCode(code string) (*structs.UserPromo, error) {
	var promo structs.UserPromo
	err := engine.DB.Where("code = ?", code).First(&promo).Error
	return &promo, err
}

func GetOwnerIDByCode(code string) (uint64, error) {
	var ownerID int64
	// Используем .Model() и .Pluck() для получения значения одной колонки
	err := engine.DB.Model(&structs.UserPromo{}).Where("code = ?", code).Pluck("owner_id", &ownerID).Error
	if err != nil {
		return 0, err
	}
	return uint64(ownerID), nil
}
