package engine

import (
	"database/sql"
	"errors"
	"florenbot/engine/structs"
	"florenbot/helpers"
	"log"
)

func CreateClan(name string, owner_name string, owner_id uint64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	res, err := tx.Exec("INSERT INTO clans (name, owner_id, owner_name) VALUES (?, ?, ?)", name, owner_id, owner_name)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Ошибка при откате транзакции: %v", err)
		}
		return err
	}

	clan_id, err := res.LastInsertId()
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Ошибка при откате транзакции: %v", err)
		}
		return err
	}

	_, err = tx.Exec("UPDATE users SET clan_id = ? WHERE id = ?", clan_id, owner_id)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Ошибка при откате транзакции: %v", err)
		}
		return err
	}

	_, err = tx.Exec("INSERT INTO clans_members (clan_id, user_id) VALUES (?, ?)", clan_id, owner_id)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Ошибка при откате транзакции: %v", err)
		}
		return err
	}

	_, err = tx.Exec("UPDATE clans_members SET role = 'owner' WHERE clan_id = ? AND user_id = ?", clan_id, owner_id)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Ошибка при откате транзакции: %v", err)
		}
		return err
	}

	return tx.Commit()
}

func GetClanOwnerID(clan_id uint64) (uint64, error) {
	var owner_id uint64
	err := DB.QueryRow("SELECT owner_id FROM clans WHERE id = ?", clan_id).Scan(&owner_id)
	return owner_id, err
}

func AddUserToClan(clan_id uint64, user_id uint64) error {
	_, err := DB.Exec("INSERT INTO clans_members (clan_id, user_id) VALUES (?, ?)", clan_id, user_id)
	return err
}

func GetUserClanID(userID uint64) (uint64, error) {
	var clanID uint64
	err := DB.QueryRow("SELECT clan_id FROM clans_members WHERE user_id = ?", userID).Scan(&clanID)
	return clanID, err
}

func DeleteMembersClan(clan_id uint64) error {
	_, err := DB.Exec("DELETE FROM clans_members WHERE clan_id = ?", clan_id)
	return err
}

func KickClanUser(clan_id uint64, user_id uint64) error {
	_, err := DB.Exec("DELETE FROM clans_members WHERE clan_id = ? AND user_id = ?", clan_id, user_id)
	return err
}

func GetUserClanRole(userID uint64) (string, error) {
	var role string
	err := DB.QueryRow("SELECT role FROM clans_members WHERE user_id = ?", userID).Scan(&role)
	return role, err
}

func GetClans() ([]structs.Clans, error) {
	query := `
    SELECT 
        c.id, 
        c.name, 
        c.owner_id, 
        c.invite_code,
        COUNT(cm.user_id) as member_count
    FROM clans c
    LEFT JOIN clans_members cm ON c.id = cm.clan_id
    GROUP BY c.id`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clans []structs.Clans
	for rows.Next() {
		var c structs.Clans
		var inviteCode sql.NullString

		if err := rows.Scan(&c.Id, &c.Name, &c.OwnerId, &inviteCode, &c.MemberCount); err != nil {
			return nil, err
		}

		if inviteCode.Valid {
			c.InviteCode = inviteCode.String
		}

		clans = append(clans, c)
	}

	return clans, nil
}

func BlockMemberClan(clan_id uint64, user_id uint64, reason string) error {
	tx, err := DB.Begin()

	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO clans_blacklist (user_id, clan_id, reason) VALUES (?, ?, ?)", user_id, clan_id, reason)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Ошибка при откате транзакции: %v", err)
		}
		return err
	}

	_, err = tx.Exec("DELETE FROM clans_members WHERE clan_id = ? AND user_id = ?", clan_id, user_id)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Ошибка при откате транзакции: %v", err)
		}
		return err
	}

	return tx.Commit()

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

func GetClanByID(id uint64) (*structs.Clans, error) {
	clan := &structs.Clans{}
	query := "SELECT id, name, owner_id FROM clans WHERE id = ?"
	err := DB.QueryRow(query, id).Scan(&clan.Id, &clan.Name, &clan.OwnerId)
	return clan, err
}

func GetClanByOwnerID(owner_id uint64) (uint64, error) {
	var id uint64
	err := DB.QueryRow("SELECT id FROM clans WHERE owner_id = ?", owner_id).Scan(&id)
	return id, err
}

func GetClan(id uint64) (*structs.Clans, error) {
	clan := &structs.Clans{}
	query := "SELECT id, name, owner_name FROM clans WHERE id = ?"
	err := DB.QueryRow(query, id).Scan(&clan.Id, &clan.Name, &clan.OwnerName)
	return clan, err
}

func CheckBlacklist(clan_id uint64, user_id uint64) error {
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM clans_blacklist WHERE user_id = ? AND clan_id = ?)", user_id, clan_id).Scan(&exists)

	if err != nil {
		return err
	}

	if exists {
		return errors.New("user_in_blacklist")
	}

	return nil
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
	var code sql.NullString

	err := DB.QueryRow("SELECT invite_code FROM clans WHERE id = ?", clan_id).Scan(&code)
	if err != nil {
		return "", err
	}

	if !code.Valid {
		return "", sql.ErrNoRows
	}

	return code.String, nil
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
		if err := tx.Rollback(); err != nil {
			log.Printf("Ошибка при откате транзакции: %v", err)
		}
		return err
	}

	newCode := helpers.GenerateCode()
	_, err = tx.Exec("UPDATE clans SET invite_code = ? WHERE id = ?", newCode, clan_id)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Ошибка при откате транзакции: %v", err)
		}
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
		if err := tx.Rollback(); err != nil {
			log.Printf("Ошибка при откате транзакции: %v", err)
		}
		return err
	}

	_, err = tx.Exec("DELETE FROM clans WHERE id = ?", id)
	if err != nil {
		if err := tx.Rollback(); err != nil {
			log.Printf("Ошибка при откате транзакции: %v", err)
		}
		return err
	}

	return tx.Commit()
}
