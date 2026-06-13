package helpers

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	engine "florenbot/engine"
	"florenbot/engine/structs"
)

// GetUserByID получает данные пользователя из Redis Hash или базы данных
func GetUserByID(tg_id uint64) (structs.User, error) {
	var user structs.User
	key := fmt.Sprintf("user:%d", tg_id)

	// 1. Читання з Redis
	data, err := engine.HGetAll(key)
	if err == nil && len(data) > 0 {
		user.Id = tg_id
		user.Username = data["username"]
		
		f, _ := strconv.ParseFloat(data["balance"], 64)
		user.Balance = float32(f)

		user.Role = data["role"]
		
		user.PromoCode = data["promocode"] // Тепер це просто рядок
		
		// Обробка вказівника ClanId
		if val, ok := data["clan_id"]; ok && val != "" && val != "0" {
			clanID, _ := strconv.ParseInt(val, 10, 64)
			user.ClanId = &clanID // Присвоюємо адресу змінної (вказівник)
		} else {
			user.ClanId = nil // Вказівник nil означає відсутність значення
		}
		
		user.NegativeReputation, _ = strconv.Atoi(data["negative_reputation"])
		user.PositiveReputation, _ = strconv.Atoi(data["positive_reputation"])
		
		return user, nil
	}

	// 2. Читання з БД (потрібні тимчасові змінні для сканування вказівників)
	var clanID sql.NullInt64
	var promoCode sql.NullString
	var role sql.NullString
	query := `SELECT id, username, balance, promocode, clan_id, negative_reputation, positive_reputation, role
              FROM users WHERE id = ?`

	err = engine.DB.QueryRow(query, tg_id).Scan(
		&user.Id, &user.Username, &user.Balance, &promoCode,
		&clanID, &user.NegativeReputation, &user.PositiveReputation, 
		&role,
	)
	if err != nil {
		if err == sql.ErrNoRows { return structs.User{}, nil }
		return structs.User{}, err
	}

	// Перенесення значень з Null-типів БД у структуру
	if promoCode.Valid { user.PromoCode = promoCode.String }
	if clanID.Valid { user.ClanId = &clanID.Int64 }
	if role.Valid { user.Role = role.String }

	// 3. Запис у Redis
	vals := map[string]interface{}{
		"username":            user.Username,
		"balance":             user.Balance,
		"promocode":           user.PromoCode,
		"clan_id":             0,
		"role":                "user",
		"negative_reputation": user.NegativeReputation,
		"positive_reputation": user.PositiveReputation,
	}
	if user.ClanId != nil {
		vals["clan_id"] = *user.ClanId
	}

	if user.Role != "" {
		vals["role"] = user.Role
	}

	_ = engine.HMSet(key, vals)
	engine.Expire(key, 10*time.Minute)

	return user, nil
}


func GetRole(user_id uint64) (string, error) {

	user, err := GetUserByID(user_id)

	if err != nil {
		return "", err
	}

	return user.Role, nil
}


// CreateUser создает пользователя и обновляет кеш
func CreateUser(user_id uint64, username string, firstName string) (structs.User, error) {
	log.Printf("[DEBUG] Создание пользователя: %d", user_id)
	query := "INSERT INTO users (id, username, first_name) VALUES (?, ?, ?)"

	_, err := engine.DB.Exec(query, user_id, username, firstName)
	if err != nil {
		return structs.User{}, err
	}

	return GetUserByID(user_id)
}

// GetUsers возвращает список всех пользователей
func GetUsers() ([]structs.User, error) {
	query := "SELECT id FROM users"
	rows, err := engine.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []structs.User
	for rows.Next() {
		var u structs.User
		if err := rows.Scan(&u.Id); err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, rows.Err()
}