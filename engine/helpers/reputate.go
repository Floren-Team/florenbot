package helpers

import (
	engine "florenbot/engine/mysql"
	"florenbot/engine/structs"
	"gorm.io/gorm"
)

// UpdateNegativeReputation устанавливает фиксированное значение негативной репутации
func UpdateNegativeReputation(userID uint64, reputation int) error {
	return engine.DB.Model(&structs.User{}).Where("id = ?", userID).Update("negative_reputation", reputation).Error
}

// UpdatePositiveReputation устанавливает фиксированное значение позитивной репутации
func UpdatePositiveReputation(userID uint64, reputation int) error {
	return engine.DB.Model(&structs.User{}).Where("id = ?", userID).Update("positive_reputation", reputation).Error
}

// AddNegativeReputation прибавляет значение к текущей негативной репутации
func AddNegativeReputation(userID uint64, reputation int) error {
	return engine.DB.Model(&structs.User{}).Where("id = ?", userID).
		Update("negative_reputation", gorm.Expr("negative_reputation + ?", reputation)).Error
}

// AddPositiveReputation прибавляет значение к текущей позитивной репутации
func AddPositiveReputation(userID uint64, reputation int) error {
	return engine.DB.Model(&structs.User{}).Where("id = ?", userID).
		Update("positive_reputation", gorm.Expr("positive_reputation + ?", reputation)).Error
}

// GetReputation вычисляет итоговую репутацию
func GetReputation(userID uint64) (int, error) {
	var user structs.User
	err := engine.DB.Select("positive_reputation", "negative_reputation").First(&user, userID).Error
	if err != nil {
		return 0, err
	}
	return user.PositiveReputation - user.NegativeReputation, nil
}

// GetNegativeReputation получает негативную репутацию
func GetNegativeReputation(userID uint64) (int, error) {
	var user structs.User
	err := engine.DB.Select("negative_reputation").First(&user, userID).Error
	return user.NegativeReputation, err
}

// GetPositiveReputation получает позитивную репутацию
func GetPositiveReputation(userID uint64) (int, error) {
	var user structs.User
	err := engine.DB.Select("positive_reputation").First(&user, userID).Error
	return user.PositiveReputation, err
}
