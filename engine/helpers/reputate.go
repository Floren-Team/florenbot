package helpers

import (
	"database/sql"
	engine "florenbot/engine"
	helpers "florenbot/helpers"
	"log"
)



func UpdateNetagiveReputation(user_id uint64, reputation int) error {
	debug_type := helpers.GetEnvBool("DEBUG", false)
	err := engine.DB.QueryRow("UPDATE users SET negative_reputation = ? WHERE id = ?", reputation, user_id).Err()
	if err != nil {
		if err == sql.ErrNoRows {
			if debug_type {
				log.Printf("Пользователь с ID %d не был найден", user_id)
				return err
			}
		}
	}
	return nil
}

func UpdatePositiveReputation_2(user_id uint64, reputation int) error {
	debug_type := helpers.GetEnvBool("DEBUG", false)
	err := engine.DB.QueryRow("UPDATE users SET positive_reputation = ? WHERE id = ?", reputation, user_id).Err()
	if err != nil {
		if err == sql.ErrNoRows {
			if debug_type {
				log.Printf("Пользователь с ID %d не был найден", user_id)
				return err
			}
		}
	}
	return nil
}

func GetReputation(user_id uint64) (int, error) {
	// 1. Получаем позитивную репутацию
	pos, err := GetPositiveReputation(user_id)
	if err != nil {
		return 0, err
	}

	// 2. Получаем негативную репутацию
	neg, err := GetNegativeReputation(user_id)
	if err != nil {
		return 0, err
	}

	// 3. Вычисляем итоговую репутацию
	// Логика: Positive - Negative (или сумма, в зависимости от вашей задумки)
	total_reputation := pos - neg

	return total_reputation, nil
}

func GetNegativeReputation(user_id uint64) (int, error) {
	var reputation int
	err := engine.DB.QueryRow("SELECT negative_reputation FROM users WHERE id = ?", user_id).Scan(&reputation)
	return reputation, err
}

func GetPositiveReputation(user_id uint64) (int, error) {
	var reputation int
	err := engine.DB.QueryRow("SELECT positive_reputation FROM users WHERE id = ?", user_id).Scan(&reputation)
	return reputation, err
}

func UpdateNegativeReputation(user_id uint64, reputation int) {
	debug_type := helpers.GetEnvBool("DEBUG", false)
	err := engine.DB.QueryRow("UPDATE users SET negative_reputation = negative_reputation + ? WHERE id = ?", reputation, user_id).Err()
	if err != nil {
		if err == sql.ErrNoRows {
			if debug_type {
				log.Printf("Пользователь с ID %d не был найден", user_id)
			}
		}
	}
}

func UpdatePositiveReputation(user_id uint64, reputation int) {
	debug_type := helpers.GetEnvBool("DEBUG", false)
	err := engine.DB.QueryRow("UPDATE users SET positive_reputation = positive_reputation + ? WHERE id = ?", reputation, user_id).Err()
	if err != nil {
		if err == sql.ErrNoRows {
			if debug_type {
				log.Printf("Пользователь с ID %d не был найден", user_id)
			}
		}
	}
}
