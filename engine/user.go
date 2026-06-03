package engine

import (
	"database/sql"
	"florenbot/engine/structs"
)

func GetUserByID(tg_id uint64) (structs.User, error) {
	var user structs.User
	query := `SELECT id, username,  balance, promocode, clan_id, negative_reputation, positive_reputation 
              FROM users WHERE id = ?`

	err := DB.QueryRow(query, tg_id).Scan(
		&user.Id,
		&user.Username,
		&user.Balance,
		&user.PromoCode,
		&user.ClanId,
		&user.NegativeReputation,
		&user.PositiveReputation,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return structs.User{}, nil
		}
		return structs.User{}, err
	}
	return user, nil
}

func CreateUser(
	user_id uint64,
	username string,
	firstName string,
) (structs.User, error) {
	query := "INSERT INTO users (id, username, first_name) VALUES (?, ?, ?)"

	_, err := DB.Exec(query, user_id, username, firstName)
	if err != nil {
		return structs.User{}, err
	}

	return GetUserByID(user_id)
}
