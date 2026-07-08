package helpers

import (
	engine "florenbot/engine/mysql"
	"florenbot/engine/structs"
	"gorm.io/gorm"
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

func GetPromocodesMemberCount() (map[string]int, error) {
    stats := make(map[string]int)

    type Result struct {
        Code  string
        Total int
    }
    var results []Result

    err := engine.DB.Model(&structs.UserPromo{}).
        Select("code, count(*) as total").
        Where("code IS NOT NULL AND code != ''"). 
        Group("code").
        Scan(&results).Error

    if err != nil {
        return nil, err
    }

    for _, res := range results {
        cleanCode := strings.TrimSpace(res.Code)
        if cleanCode != "" {
            stats[cleanCode] = res.Total
        }
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
