package engine

import (
	"database/sql"
	"florenbot/engine/structs"
)

func ActivateCode(id uint64, code string) error {
	debug("Активация кода: %s для пользователя %d", code, id)
	_, err := DB.Exec("UPDATE users SET promocode = ? WHERE id = ?", code, id)
	return err
}

func CreateCode(code string, amount int) error {
	debug("Создание кода: %s, сумма: %d", code, amount)
	_, err := DB.Exec("INSERT INTO promocodes (code, amount) VALUES (?, ?)", code, amount)
	return err
}

func GetUserCode(userID uint64) (string, error) {
	var code sql.NullString
	err := DB.QueryRow("SELECT promocode FROM users WHERE id = ?", userID).Scan(&code)
	if err != nil {
		return "", err
	}
	return code.String, nil
}

func DeleteCode(code string) error {
	debug("Удаление кода: %s", code)
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM promocodes WHERE code = ?", code)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("UPDATE users SET promocode = NULL WHERE promocode = ?", code)
	debug("Результат обнуления промокода у пользователей: err=%v", err)

	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func GetPromocodesUser(id uint64) ([]model.UserPromo, error) {
	code, err := GetUserCode(id)
	if err != nil || code == "" {
		return nil, err
	}

	amount, _ := GetCode(code)

	// Возвращаем список из одного элемента
	return []model.UserPromo{{Code: code, Amount: amount}}, nil
}

func GetCode(code string) (int, error) {
	debug("Проверка кода: %s", code)
	var amount int
	err := DB.QueryRow("SELECT amount FROM promocodes WHERE code = ?", code).Scan(&amount)
	return amount, err
}
