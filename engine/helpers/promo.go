package helpers

import (
	"database/sql"
	engine "florenbot/engine/mysql"
	"florenbot/engine/structs"
	"log"
)

func ActivateCode(id uint64, code string) error {
	engine.Debug("Активация кода: %s для пользователя %d", code, id)
	_, err := engine.DB.Exec("UPDATE users SET promocode = ? WHERE id = ?", code, id)
	return err
}

func CreateCode(code string, amount int, owner_id uint64) error {
	engine.Debug("Создание кода: %s, сумма: %d", code, amount)
	_, err := engine.DB.Exec("INSERT INTO promocodes (code, amount, owner_id) VALUES (?, ?, ?)", code, amount, owner_id)
	return err
}

func GetPromocodesMemberCount() (map[string]int, error) {
	stats := make(map[string]int)
	query := `SELECT p.code, COUNT(u.id) FROM promocodes p LEFT JOIN users u ON p.code = u.promocode GROUP BY p.code`

	rows, err := engine.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var code string
		var count int
		if err := rows.Scan(&code, &count); err != nil {
			return nil, err
		}
		stats[code] = count
	}
	return stats, nil
}

func GetUserCode(userID uint64) (string, error) {
	var code sql.NullString
	err := engine.DB.QueryRow("SELECT promocode FROM users WHERE id = ?", userID).Scan(&code)
	if err != nil {
		return "", err
	}
	return code.String, nil
}

func DeleteCode(code string) error {
	engine.Debug("Удаление кода: %s", code)
	tx, err := engine.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM promocodes WHERE code = ?", code)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Ошибка при откате транзакции: %v", err)
		}
		return err
	}

	_, err = tx.Exec("UPDATE users SET promocode = NULL WHERE promocode = ?", code)
	engine.Debug("Результат обнуления промокода у пользователей: err=%v", err)

	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Ошибка при откате транзакции: %v", err)
		}
		return err
	}

	return tx.Commit()
}

func GetPromocodesUser(id uint64) ([]structs.UserPromo, error) {
	code, err := GetUserCode(id)
	if err != nil || code == "" {
		return nil, err
	}

	amount, _ := GetCode(code)

	// Возвращаем список из одного элемента
	return []structs.UserPromo{{Code: code, Amount: amount}}, nil
}

func UpdatePromo(code string, amount float64, user_id uint64) error {
	engine.Debug("Обновление кода: %s, сумма: %.2f", code, amount)
	tx, err := engine.DB.Begin()

	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE promocodes SET amount = ?, code = ? WHERE owner_id = ?", amount, code, user_id)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Ошибка при откате транзакции: %v", err)
		}
		return err
	}

	_, err = tx.Exec("UPDATE users SET promocode = ? WHERE id = ?", code, user_id)

	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Ошибка при откате транзакции: %v", err)
		}
		return err
	}

	return tx.Commit()
}

func GetCode(code string) (float64, error) {
	engine.Debug("Проверка кода: %s", code)
	var amount float64
	err := engine.DB.QueryRow("SELECT amount FROM promocodes WHERE code = ?", code).Scan(&amount)
	return amount, err
}
