package engine

import (
	"florenbot/engine/model"
	"florenbot/helpers"
	"log"
)

func CreateClan(name string, owner_id uint64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	res, err := tx.Exec("INSERT INTO clans (name, owner_id) VALUES (?, ?)", name, owner_id)
	if err != nil {
		tx.Rollback()
		return err
	}

	clan_id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("UPDATE users SET clan_id = ? WHERE id = ?", clan_id, owner_id)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("INSERT INTO clans_members (clan_id, user_id) VALUES (?, ?)", clan_id, owner_id)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func GetUserClanID(userID uint64) (uint64, error) {
	var clanID uint64
	err := DB.QueryRow("SELECT clan_id FROM clans_members WHERE user_id = ?", userID).Scan(&clanID)
	return clanID, err
}

func GetClanMemberCount(clanID int64) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM clans_members WHERE clan_id = ?"

	err := DB.QueryRow(query, clanID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func GetClanByID(id uint64) (*model.Clans, error) {
	clan := &model.Clans{}
	query := "SELECT id, name, owner_id FROM clans WHERE id = ?"
	err := DB.QueryRow(query, id).Scan(&clan.Id, &clan.Name, &clan.OwnerId)
	return clan, err
}

func GetClanByOwnerID(owner_id uint64) (uint64, error) {
	var id uint64
	err := DB.QueryRow("SELECT id FROM clans WHERE owner_id = ?", owner_id).Scan(&id)
	return id, err
}

func GetClan(id uint64) (*model.Clans, error) {
	clan := &model.Clans{}
	query := "SELECT id, name, owner_id FROM clans WHERE id = ?"
	err := DB.QueryRow(query, id).Scan(&clan.Id, &clan.Name, &clan.OwnerId)
	return clan, err
}

func JoinClan(clan_id uint64, user_id uint64) error {
	_, err := DB.Exec("INSERT INTO clans_members (clan_id, user_id) VALUES (?, ?)", clan_id, user_id)
	return err
}

func LeaveClan(clan_id uint64, user_id uint64) error {
	query := "DELETE FROM clans_members WHERE clan_id = ? AND user_id = ?"
	result, err := DB.Exec(query, clan_id, user_id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Print("Запись не найдена или не была удалена")
	}

	return nil
}

func GetInviteCodeClan(code string) (uint64, error) {
	var clanID uint64
	err := DB.QueryRow("SELECT id FROM clans WHERE invite_code = ?", code).Scan(&clanID)
	if err != nil {
		return 0, err
	}
	return clanID, nil
}

func GetClanInviteCode(clan_id uint64) (string, error) {
	var code string
	err := DB.QueryRow("SELECT invite_code FROM clans WHERE id = ?", clan_id).Scan(&code)
	return code, err
}

func DeleteInviteCode(clan_id uint64) error {
	_, err := DB.Exec("UPDATE clans SET invite_code = NULL WHERE id = ?", clan_id)
	return err
}

func RevokeInviteCode(clan_id uint64) error {
	tx, err := DB.Begin()

	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE clans SET invite_code = NULL WHERE id = ?", clan_id)
	if err != nil {
		tx.Rollback()
		return err
	}

	newCode := helpers.GenerateCode()
	_, err = tx.Exec("UPDATE clans SET invite_code = ? WHERE id = ?", newCode, clan_id)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func CreateInviteCode(clan_id uint64, code string) error {
	_, err := DB.Exec("UPDATE clans SET invite_code = ? WHERE id = ?", code, clan_id)
	return err
}

func GetClanMember(clan_id uint64, user_id uint64) (uint16, error) {
	var id uint16
	err := DB.QueryRow("SELECT user_id FROM clans_members WHERE clan_id = ? AND user_id = ?", clan_id, user_id).Scan(&id)
	return id, err
}

func DeleteClan(id uint64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE users SET clan_id = NULL WHERE clan_id = ?", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("DELETE FROM clans WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
