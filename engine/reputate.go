package engine

import (
	"database/sql"
	"log"
)

func ParseEnvBool(key string, defaultValue bool) bool {
	value := getEnv(key, "")
	if value == "" {
		return defaultValue
	}
	return value == "true"
}

func UpdateReputation(user_id uint64, reputation int) error {
	debug_type := ParseEnvBool("DEBUG", false)
	err := DB.QueryRow("UPDATE users SET reputation = reputation + ? WHERE id = ?", reputation, user_id).Err()
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
	var reputation int
	err := DB.QueryRow("SELECT reputation FROM users WHERE id = ?", user_id).Scan(&reputation)
	return reputation, err
}

func GetNegativeReputation(user_id uint64) (int, error) {
	var reputation int
	err := DB.QueryRow("SELECT negative_reputation FROM users WHERE id = ?", user_id).Scan(&reputation)
	return reputation, err
}

func GetPositiveReputation(user_id uint64) (int, error) {
	var reputation int
	err := DB.QueryRow("SELECT positive_reputation FROM users WHERE id = ?", user_id).Scan(&reputation)
	return reputation, err
}

func UpdateNegativeReputation(user_id uint64, reputation int) {
	debug_type := ParseEnvBool("DEBUG", false)
	err := DB.QueryRow("UPDATE users SET negative_reputation = negative_reputation + ? WHERE id = ?", reputation, user_id).Err()
	if err != nil {
		if err == sql.ErrNoRows {
			if debug_type {
				log.Printf("Пользователь с ID %d не был найден", user_id)
			}
		}
	}
}

func UpdatePositiveReputation(user_id uint64, reputation int) {
	debug_type := ParseEnvBool("DEBUG", false)
	err := DB.QueryRow("UPDATE users SET positive_reputation = positive_reputation + ? WHERE id = ?", reputation, user_id).Err()
	if err != nil {
		if err == sql.ErrNoRows {
			if debug_type {
				log.Printf("Пользователь с ID %d не был найден", user_id)
			}
		}
	}
}
