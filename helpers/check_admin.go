package helpers

import (
	"florenbot/engine/structs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// IsUserAdmin проверяет права администратора для объекта роли
func IsUserAdmin(role *structs.Role) bool {
	// Если роль nil, прав нет
	if role == nil {
		return false
	}
	return role.BaseShort == "admin" || role.BaseShort == "owner" || role.BaseShort == "creator"
}

func IsUserOwner(role *structs.Role) bool {
	if role == nil {
		return false
	}
	return role.BaseShort == "owner"
}

func IsUserCreator(role *structs.Role) bool {
	if role == nil {
		return false
	}
	return role.BaseShort == "creator"
}

func IsUserOwnerOrCreator(role *structs.Role) bool {
	if role == nil {
		return false
	}
	return role.BaseShort == "owner" || role.BaseShort == "creator"
}

func IsCreator(bot *tgbotapi.BotAPI, chatID int64, userID int64) (bool, error) {
	// Створюємо запит на отримання статусу учасника
	config := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: chatID,
			UserID: userID,
		},
	}

	// Отримуємо інформацію про учасника
	member, err := bot.GetChatMember(config)
	if err != nil {
		return false, err
	}

	// Перевіряємо, чи статус дорівнює "creator"
	return member.Status == "creator", nil
}