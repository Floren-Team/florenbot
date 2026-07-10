package helpers

import (
	engine "florenbot/engine/mysql"
	"florenbot/engine/structs"
)

func GetRolesByChatID(chat_id uint64) ([]structs.Role, error) {
	var roles []structs.Role
	err := engine.DB.Where("chat_id = ?", chat_id).
		Order("priority DESC").
		Find(&roles).Error
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

func InitDefaultRoles(chat_id uint64) error {
	// Определяем стандартные роли с их приоритетами
	// Чем выше число, тем выше уровень доступа (права на снятие/изменение)
	defaultRoles := []structs.Role{
		{Name: "Создатель", ShortName: "creator", BaseShort: "creator", ChatID: chat_id, Priority: 100},
		{Name: "Владелец", ShortName: "owner", BaseShort: "owner", ChatID: chat_id, Priority: 90},
		{Name: "Администратор", ShortName: "admin", BaseShort: "admin", ChatID: chat_id, Priority: 50},
		{Name: "Модератор", ShortName: "moderator", BaseShort: "moderator", ChatID: chat_id, Priority: 20},
		{Name: "Участник", ShortName: "member", BaseShort: "member", ChatID: chat_id, Priority: 10},
	}

	for _, role := range defaultRoles {
		var count int64
		// Проверяем наличие роли в базе данных
		engine.DB.Model(&structs.Role{}).
			Where("short_name = ? AND chat_id = ?", role.ShortName, chat_id).
			Count(&count)

		// Если роль не найдена, создаем её
		if count == 0 {
			if err := engine.DB.Create(&role).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func AddMemberRole(user_id uint64, baseShort string, chat_id uint64) error {
	// 1. Начало транзакции
	tx := engine.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 2. Поиск ID роли в текущем чате
	var role_id uint64
	if err := tx.Model(&structs.Role{}).
		Where("short_name = ? AND chat_id = ?", baseShort, chat_id).
		Select("id").
		Scan(&role_id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 3. Обновление или создание записи участника в таблице Members
	// Мы работаем ТОЛЬКО с таблицей участников
	if err := tx.Where(structs.Member{ChatID: chat_id, UserID: user_id}).
		Assign(structs.Member{RoleID: role_id}).
		FirstOrCreate(&structs.Member{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// --- БЛОК ОБНОВЛЕНИЯ USERS УДАЛЕН, ТАК КАК ПОЛЯ ТАМ БОЛЬШЕ НЕТ ---

	// 4. Фиксация транзакции
	if err := tx.Commit().Error; err != nil {
		return err
	}

	// 5. ОБНОВЛЕНИЕ КЕША
	// Поскольку мы изменили только таблицу members, а в кеше пользователя
	// роль больше не хранится, нам нужно просто обновить кеш пользователя
	// на случай, если там есть другие актуальные данные.
	updatedUser, err := GetUserByID(user_id)
	if err == nil {
		UpdateUserCache(updatedUser)
	}

	return nil
}

// SetMemberRole устанавливает роль пользователю в конкретном чате
func SetMemberRole(userID uint64, roleID uint64, chatID uint64) error {
	// Используем транзакцию для атомарности
	tx := engine.DB.Begin()

	// Обновляем или создаем запись участника
	member := structs.Member{ChatID: chatID, UserID: userID}
	if err := tx.Where(member).Assign(structs.Member{RoleID: roleID}).FirstOrCreate(&member).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	// После изменения роли в БД, нужно принудительно обновить кеш пользователя
	// так как профиль пользователя может содержать информацию о текущем чате или правах
	user, err := GetUserByID(userID)
	if err == nil {
		UpdateUserCache(user)
	}

	return nil
}

func CreateRole(name string, short_name string, base_short string, chat_id uint64, priority int) error {
	role := structs.Role{
		Name:      name,
		ShortName: short_name,
		BaseShort: base_short,
		ChatID:    chat_id,
		Priority:  priority,
	}
	return engine.DB.Save(&role).Error
}

func GetRoleByID(roleID uint64, chat_id uint64) (structs.Role, error) {
	var role structs.Role
	err := engine.DB.Where("id = ? AND chat_id = ?", roleID, chat_id).First(&role).Error
	return role, err
}

func GetRoleByBaseShort(short_name string, chat_id uint64) (structs.Role, error) {
	var role structs.Role
	err := engine.DB.Where("base_short = ? AND chat_id = ?", short_name, chat_id).First(&role).Error
	return role, err
}

func GetRoleByShortName(short_name string, chat_id uint64) (structs.Role, error) {
	var role structs.Role
	err := engine.DB.Where("short_name = ? AND chat_id = ?", short_name, chat_id).First(&role).Error
	return role, err
}

// GetUsersByRoleID — универсальная функция для получения пользователей по ID роли
// GetUsersByRoleID — исправленная функция с использованием JOIN
func GetUsersByRoleID(role_id uint64, chat_id uint64) ([]structs.User, error) {
	var users []structs.User

	// Используем Join, чтобы обратиться к столбцу role_id в таблице chat_members
	err := engine.DB.Table("users").
		Select("users.*").
		Joins("JOIN chat_members ON chat_members.user_id = users.id").
		Where("chat_members.role_id = ? AND chat_members.chat_id = ?", role_id, chat_id).
		Where("users.deleted_at IS NULL").
		Find(&users).Error

	return users, err
}

// GetAdminsUsers — получение всех администраторов в чате
func GetAdminsUsers(chat_id uint64) ([]structs.User, error) {
	// 1. Сначала ищем саму роль администратора
	var role structs.Role
	err := engine.DB.Where("base_short = ? AND chat_id = ?", "admin", chat_id).First(&role).Error
	if err != nil {
		return nil, err
	}

	// 2. Получаем пользователей по ID этой роли
	return GetUsersByRoleID(role.ID, chat_id)
}

// GetModeratorsUsers — получение всех модераторов в чате
func GetModeratorsUsers(chat_id uint64) ([]structs.User, error) {
	// 1. Сначала ищем саму роль модератора
	var role structs.Role
	err := engine.DB.Where("base_short = ? AND chat_id = ?", "moderator", chat_id).First(&role).Error
	if err != nil {
		return nil, err
	}

	// 2. Получаем пользователей по ID этой роли
	return GetUsersByRoleID(role.ID, chat_id)
}

func RemoveRole(user_id uint64, chat_id uint64) error {
	// Нахождение ID роли, которую нужно присвоить
	role, err := GetRoleByShortName("member", chat_id)
	if err != nil {
		return err
	}
	role_id := role.ID

	data := structs.Member{ChatID: chat_id, UserID: user_id, RoleID: role_id}

	// Обновление записи

	err = engine.DB.Where("chat_id = ? AND user_id = ?", chat_id, user_id).Updates(data).Error
	return err

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
