package engine

import (
	"database/sql"
	"florenbot/engine/structs"
)

func GetUserByID(tg_id uint64) (model.User, error) {
	var user model.User
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
			return model.User{}, nil
		}
		return model.User{}, err
	}
	return user, nil
}
