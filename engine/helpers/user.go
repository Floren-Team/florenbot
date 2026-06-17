package helpers

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"
	engine "florenbot/engine/mysql"
	cache "florenbot/engine/cache"
	"florenbot/engine/structs"
)

// GetUserByID получает данные пользователя из Redis Hash или базы данных
func GetUserByID(tg_id uint64) (structs.User, error) {
	var user structs.User
	key := fmt.Sprintf("user:%d", tg_id)

	data, err := cache.HGetAll(key)
	if err == nil && len(data) > 0 {
		user.Id = tg_id
		user.Username = data["username"]

		f, _ := strconv.ParseFloat(data["balance"], 64)
		user.Balance = float32(f)

		user.Role = data["role"]
		user.PromoCode = data["promocode"]

		user.Losses, _ = strconv.Atoi(data["losses"])
		user.Wins, _ = strconv.Atoi(data["wins"])
		if e, err := strconv.ParseFloat(data["euro"], 64); err == nil {
			user.Euro = float32(e)
		}

		if val, ok := data["clan_id"]; ok && val != "" && val != "0" {
			clanID, _ := strconv.ParseInt(val, 10, 64)
			user.ClanId = &clanID
		}

		user.NegativeReputation, _ = strconv.Atoi(data["negative_reputation"])
		user.PositiveReputation, _ = strconv.Atoi(data["positive_reputation"])

		return user, nil
	}

	var clanID sql.NullInt64
	var promoCode sql.NullString
	var role sql.NullString
	
	query := `SELECT id, username, balance, promocode, clan_id, negative_reputation, positive_reputation, role,
              losses, wins, euro
              FROM users WHERE id = ?`

	err = engine.DB.QueryRow(query, tg_id).Scan(
		&user.Id, &user.Username, &user.Balance, &promoCode,
		&clanID, &user.NegativeReputation, &user.PositiveReputation,
		&role, &user.Losses, &user.Wins, &user.Euro,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return structs.User{}, nil
		}
		return structs.User{}, err
	}

	if promoCode.Valid { user.PromoCode = promoCode.String }
	if clanID.Valid { user.ClanId = &clanID.Int64 }
	if role.Valid { user.Role = role.String }

	// 3. Запис у кеш
	vals := map[string]interface{}{
		"username":            user.Username,
		"balance":             user.Balance,
		"promocode":           user.PromoCode,
		"clan_id":             0,
		"role":                "user",
		"negative_reputation": user.NegativeReputation,
		"positive_reputation": user.PositiveReputation,
		"losses":              user.Losses,
		"wins":                user.Wins,
		"euro":                user.Euro,
	}
	if user.ClanId != nil { vals["clan_id"] = *user.ClanId }
	if user.Role != "" { vals["role"] = user.Role }

	_ = cache.HMSet(key, vals)
	cache.Expire(key, 10*time.Minute)

	return user, nil
}

func AddUserToLosses(user_id uint64) error {
	_, err := engine.DB.Exec("UPDATE users SET losses = losses + 1 WHERE id = ?", user_id)
	if err != nil {
		return err
	}

	return nil
}

func AddUserToWins(user_id uint64) error {
	_, err := engine.DB.Exec("UPDATE users SET wins = wins + 1 WHERE id = ?", user_id)
	if err != nil {
		return err
	}
	return nil
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
