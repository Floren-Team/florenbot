package helpers

import (
	"errors"
	engine "florenbot/engine/mysql"
	"florenbot/engine/structs"
	"gorm.io/gorm"
	// Используем пакет для генерации кодов
	"florenbot/helpers"
)

// --- Основные операции ---

// CreateClan создает клан, владельца и участника в транзакции
func CreateClan(name string, ownerName string, ownerID uint64) error {
	return engine.DB.Transaction(func(tx *gorm.DB) error {
		clan := structs.Clan{Name: name, OwnerID: int64(ownerID), OwnerName: ownerName}
		if err := tx.Create(&clan).Error; err != nil {
			return err
		}
		if err := tx.Model(&structs.User{}).Where("id = ?", ownerID).Update("clan_id", clan.ID).Error; err != nil {
			return err
		}
		member := structs.ClanMember{ClanID: uint(clan.ID), UserID: int64(ownerID), Role: "owner"}
		return tx.Create(&member).Error
	})
}

// KickClanUser исключает пользователя из клана
func KickClanUser(clanID, userID uint64) error {
	return engine.DB.Where("clan_id = ? AND user_id = ?", clanID, userID).Delete(&structs.ClanMember{}).Error
}

// JoinClan добавляет пользователя в клан с проверкой черного списка
func JoinClan(clanID, userID uint64) error {
	return engine.DB.Transaction(func(tx *gorm.DB) error {
		if err := CheckBlacklist(clanID, userID); err != nil {
			return err
		}
		member := structs.ClanMember{ClanID: uint(clanID), UserID: int64(userID), Role: "member"}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}
		return tx.Model(&structs.User{}).Where("id = ?", userID).Update("clan_id", clanID).Error
	})
}

// DeleteClan удаляет клан и очищает clan_id у участников
func DeleteClan(id uint64) error {
    return engine.DB.Transaction(func(tx *gorm.DB) error {
        if err := tx.Model(&structs.User{}).Where("clan_id = ?", id).Update("clan_id", nil).Error; err != nil {
            return err
        }

        if err := tx.Where("clan_id = ?", id).Delete(&structs.ClanMember{}).Error; err != nil {
            return err
        }

        return tx.Delete(&structs.Clan{}, id).Error
    })
}

// --- Получение данных (Getters) ---

func GetClanByID(id uint64) (*structs.Clan, error) {
	var clan structs.Clan
	err := engine.DB.First(&clan, id).Error
	if err != nil {
		return nil, err
	}
	return &clan, nil
}
func GetClanByOwnerID(ownerID uint64) (*structs.Clan, error) {
	var clan structs.Clan
	err := engine.DB.Where("owner_id = ?", ownerID).First(&clan).Error
	return &clan, err
}

func GetClanOwnerID(clanID uint64) (uint64, error) {
	var clan structs.Clan
	err := engine.DB.Select("owner_id").First(&clan, clanID).Error
	return uint64(clan.OwnerID), err
}

func GetUserClanID(userID uint64) (uint64, error) {
	var member structs.ClanMember
	err := engine.DB.Select("clan_id").Where("user_id = ?", userID).First(&member).Error
	return uint64(member.ClanID), err
}

func GetUserClanRole(userID uint64) (string, error) {
	var member structs.ClanMember
	err := engine.DB.Select("role").Where("user_id = ?", userID).First(&member).Error
	return member.Role, err
}

func GetClanMemberCount(clanID uint64) (int64, error) {
	var count int64
	err := engine.DB.Model(&structs.ClanMember{}).Where("clan_id = ?", clanID).Count(&count).Error
	return count, err
}

func GetClans() ([]structs.Clan, error) {
	var clans []structs.Clan
	err := engine.DB.Table("clans").
		Select("clans.*, COUNT(clan_members.user_id) as member_count").
		Joins("LEFT JOIN clan_members ON clan_members.clan_id = clans.id").
		Group("clans.id").
		Find(&clans).Error
	return clans, err
}

// --- Работа с кодами приглашений ---

func GetClanByInviteCode(code string) (*structs.Clan, error) {
	var clan structs.Clan
	err := engine.DB.Where("invite_code = ?", code).First(&clan).Error
	return &clan, err
}

func GetClanInviteCode(clanID uint64) (string, error) {
	var clan structs.Clan
	err := engine.DB.Select("invite_code").First(&clan, clanID).Error
	if err != nil {
		return "", err
	}
	return clan.InviteCode, nil
}

func CreateInviteCode(clanID uint64, code string) error {
	return engine.DB.Model(&structs.Clan{}).Where("id = ?", clanID).Update("invite_code", code).Error
}

func DeleteInviteCode(clanID uint64) error {
	return engine.DB.Model(&structs.Clan{}).Where("id = ?", clanID).Update("invite_code", nil).Error
}

func RevokeInviteCode(clanID uint64) error {
	return engine.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&structs.Clan{}).Where("id = ?", clanID).Update("invite_code", nil).Error; err != nil {
			return err
		}
		newCode := helpers.GenerateCode()
		return tx.Model(&structs.Clan{}).Where("id = ?", clanID).Update("invite_code", newCode).Error
	})
}

// --- Управление участниками ---

func AddUserToClan(clanID, userID uint64) error {
	member := structs.ClanMember{ClanID: uint(clanID), UserID: int64(userID), Role: "member"}
	return engine.DB.Create(&member).Error
}

func LeaveClan(clanID, userID uint64) error {
	return engine.DB.Where("clan_id = ? AND user_id = ?", clanID, userID).Delete(&structs.ClanMember{}).Error
}

func BlockMemberClan(clanID, userID uint64, reason string) error {
	return engine.DB.Transaction(func(tx *gorm.DB) error {
		blacklist := structs.ClanBlacklist{ClanID: uint(clanID), UserID: int64(userID), Reason: reason}
		if err := tx.Create(&blacklist).Error; err != nil {
			return err
		}
		return tx.Where("clan_id = ? AND user_id = ?", clanID, userID).Delete(&structs.ClanMember{}).Error
	})
}

func CheckBlacklist(clanID uint64, userID uint64) error {
	var count int64
	engine.DB.Model(&structs.ClanBlacklist{}).Where("clan_id = ? AND user_id = ?", clanID, userID).Count(&count)
	if count > 0 {
		return errors.New("user_in_blacklist")
	}
	return nil
}
