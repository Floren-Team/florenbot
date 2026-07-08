package helpers

import (
	engine "florenbot/engine/mysql" // Импорт вашего пакета с DB
	"florenbot/engine/structs"
	"log"
)

// CreateChat создает запись чата
func CreateChat(chat structs.Chat) error {
	// GORM автоматически вставляет данные из структуры
	result := engine.DB.Create(&chat)
	if result.Error != nil {
		log.Printf("Ошибка при создании чата: %v", result.Error)
		return result.Error
	}
	return nil
}

// GetChatById находит чат по его ID
func GetChatById(id int64) (*structs.Chat, error) {
	var chat structs.Chat
	// First вернет ошибку, если запись не найдена
	err := engine.DB.First(&chat, id).Error
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

// GetChats получает все записи чатов
func GetChats() ([]structs.Chat, error) {
	var chats []structs.Chat
	// Find автоматически сканирует все записи в срез (slice)
	err := engine.DB.Find(&chats).Error
	if err != nil {
		log.Printf("Ошибка при получении списка чатов: %v", err)
		return nil, err
	}
	return chats, nil
}

// DeleteChat удаляет чат по ID
func DeleteChat(id int64) error {
	// Delete принимает структуру или ID.
	// Мы передаем пустую структуру с ID для корректного выполнения.
	result := engine.DB.Delete(&structs.Chat{}, id)
	if result.Error != nil {
		log.Printf("Ошибка при удалении чата: %v", result.Error)
		return result.Error
	}
	return nil
}
