package helpers

import (
	engine "florenbot/engine/mysql"
	"florenbot/engine/structs"
)

func GetRolesByChatID(chat_id uint64) ([]structs.Role, error) {
	var roles []structs.Role
	err := engine.DB.Where("chat_id = ?", chat_id).Find(&roles).Error
	return roles, err
}

func EditRole(role structs.Role) ([]structs.Role, error) {
	// 1. Обновляем конкретную запись в базе данных.
	// Используем ID и ChatID, чтобы GORM знал, какую именно строку нужно обновить,
	// и чтобы предотвратить изменение ролей из других чатов (безопасность).
	err := engine.DB.Model(&structs.Role{}).
		Where("id = ? AND chat_id = ?", role.ID, role.ChatID).
		Updates(map[string]interface{}{
			"name":       role.Name,
			"short_name": role.ShortName,
		}).Error

	// Если при обновлении произошла ошибка, возвращаем nil вместо списка и саму ошибку.
	if err != nil {
		return nil, err
	}

	// 2. Возвращаем обновленный список всех ролей для этого чата.
	return GetRolesByChatID(role.ChatID)
}

func GetUsersByRole(roleID uint64, chatID uint64) ([]structs.User, error) {
	var users []structs.User

	// Используем метод Table("users"), чтобы указать основную таблицу,
	// из которой мы хотим получить данные (структуры User).
	err := engine.DB.Table("users").
		// Выбираем все поля из таблицы users (users.*),
		// чтобы GORM заполнил структуру structs.User данными из базы.
		Select("users.*").
		// Выполняем JOIN (объединение) с таблицей members,
		// сопоставляя ID пользователя в обеих таблицах.
		Joins("JOIN chat_members ON chat_members.user_id = users.id").
		// Фильтруем результат по конкретной роли и чату,
		// которые хранятся именно в таблице связей (members).
		Where("chat_members.role_id = ? AND chat_members.chat_id = ?", roleID, chatID).
		Find(&users).Error

	return users, err
}

func SetRole(user_id uint64, role_id uint64, chat_id uint64) error {
	member := structs.Member{
		ChatID: chat_id,
		UserID: user_id,
		RoleID: role_id,
	}
	return engine.DB.Save(&member).Error
}

func CreateRole(name string, short_name string, chat_id uint64) error {
	role := structs.Role{
		Name:      name,
		ShortName: short_name,
		ChatID:    chat_id,
	}
	return engine.DB.Save(&role).Error
}

func DeleteRoleByID(roleID uint64, chat_id uint64) error {
	var role structs.Role
	err := engine.DB.Where("id = ? AND chat_id = ?", roleID, chat_id).First(&role).Error
	if err != nil {
		return err
	}
	return engine.DB.Delete(&role).Error
}

func GetMemberRole(userID uint64, chatID uint64) (structs.Role, error) {
	var role structs.Role

	err := engine.DB.Table("chat_members").
		Select("roles.*").
		Joins("JOIN roles ON roles.id = chat_members.role_id").
		Where("chat_members.user_id = ? AND chat_members.chat_id = ?", userID, chatID).
		Scan(&role).Error

	return role, err
}
