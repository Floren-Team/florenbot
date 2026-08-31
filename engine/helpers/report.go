package helpers

import (
	engine "florenbot/engine/mysql"
	"florenbot/engine/structs"
	"log"
)

// CreateReport создает новый отчет
func CreateReport(userID uint64, text string) error {
	report := structs.Report{
		UserID: &userID, // Используем указатель, так как в структуре это *int64
		Text:   text,
	}

	result := engine.DB.Create(&report)
	if result.Error != nil {
		log.Printf("Ошибка при создании отчета: %v", result.Error)
		return result.Error
	}
	return nil
}

// GetReports получает все отчеты
func GetReports() ([]structs.Report, error) {
	var reports []structs.Report

	// Find автоматически выбирает все записи из таблицы
	err := engine.DB.Find(&reports).Error
	if err != nil {
		log.Printf("Ошибка при получении списка отчетов: %v", err)
		return nil, err
	}
	return reports, nil
}

// HasReport проверяет, есть ли отчет у данного пользователя
func HasReport(userID uint64) (bool, error) {
	var count int64

	// Count — самый быстрый способ проверить существование записи в GORM
	err := engine.DB.Model(&structs.Report{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// DeleteReport удаляет отчет пользователя
func DeleteReport(userID uint64) error {
	// Delete удаляет записи, соответствующие условию
	result := engine.DB.Where("user_id = ?", userID).Delete(&structs.Report{})
	if result.Error != nil {
		log.Printf("Ошибка при удалении отчета: %v", result.Error)
		return result.Error
	}
	return nil
}
