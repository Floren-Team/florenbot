package mysql

import (
	"fmt"
	"log"
	"os"

	"florenbot/engine/structs"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// DB — глобальная переменная для работы с базой данных
var DB *gorm.DB

// Debug выводит отладочное сообщение в консоль
func Debug(message string, args ...interface{}) {
	fmt.Printf("[DEBUG] "+message+"\n", args...)
}

// ConnectDB устанавливает соединение с базой данных и выполняет миграции
func ConnectDB() {
	// 1. Загружаем переменные окружения из .env, игнорируем ошибку, если файла нет (для Docker)
	if err := godotenv.Load(); err != nil {
		log.Println("Внимание: .env файл не найден (используем системные переменные окружения)")
	}

	// 2. Формируем строку подключения (DSN) для MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	// 3. Подключаемся и записываем результат в глобальную переменную DB
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: false,
	})
	if err != nil {
		log.Panicf("Не удалось подключиться к базе данных: %v", err)
	}

	// 4. Автоматически запускаем миграции таблиц после успешного подключения
	if err := Migrate(DB); err != nil {
		log.Panicf("Ошибка при выполнении миграций: %v", err)
	}

	fmt.Println("Подключение к БД и миграции выполнены успешно!")
}

// Migrate отвечает за создание и обновление структуры таблиц в БД
// func Migrate(db *gorm.DB) error {
// 	err := db.AutoMigrate(
// 		&structs.Chat{},
// 		&structs.Clan{},
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	err = db.AutoMigrate(
// 		&structs.Role{},
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	err = db.AutoMigrate(
// 		&structs.Member{},
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	err = db.AutoMigrate(
// 		&structs.User{},
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	return db.AutoMigrate(
// 		&structs.UserPromo{},
// 		&structs.Report{},
// 		&structs.Ban{},
// 		&structs.ClanMember{},
// 		&structs.ClanBlacklist{},
// 	)
// }

// Migrate выполняет автоматическую миграцию таблиц и инициализацию данных
func Migrate(db *gorm.DB) error {
    // 1. Создаем независимые таблицы (фундамент)
    err := db.AutoMigrate(
        &structs.Chat{},
        &structs.Clan{},
    )
    if err != nil {
        return err
    }

    // 2. Создаем таблицу ролей
    err = db.AutoMigrate(&structs.Role{})
    if err != nil {
        return err
    }
    
  

    // 3. Создаем таблицы, которые зависят от ролей или чатов
    err = db.AutoMigrate(
        &structs.Member{},
        &structs.User{},
    )
    if err != nil {
        return err
    }

    // 4. Создаем все остальные зависимые таблицы
    return db.AutoMigrate(
		&structs.Role{},
        &structs.UserPromo{},
        &structs.Report{},
        &structs.Ban{},
        &structs.ClanMember{},
        &structs.ClanBlacklist{},
    )
}

// initDefaultRoles заполняет базу данных начальными ролями, если их еще нет




// GetUserBalance получает баланс пользователя по его ID
func GetUserBalance(userId uint64) (float64, error) {
	var balance float64
	result := DB.Model(&structs.User{}).Where("id = ?", userId).Pluck("balance", &balance)
	if result.Error != nil {
		return 0, result.Error
	}
	// Если записей нет, возвращаем ошибку
	if result.RowsAffected == 0 {
		return 0, fmt.Errorf("пользователь с ID %d не найден", userId)
	}
	return balance, nil
}

// UpdateBalance изменяет баланс пользователя (добавляет или вычитает сумму)
func UpdateBalance(id uint64, amount float64) error {
	result := DB.Model(&structs.User{}).Where("id = ?", id).Update("balance", gorm.Expr("balance + ?", amount))
	return result.Error
}

// CloseDB корректно закрывает соединение с базой данных
func CloseDB() {
	if DB == nil {
		return
	}
	sqlDB, err := DB.DB()
	if err != nil {
		log.Printf("Не удалось получить экземпляр БД для закрытия: %v", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Printf("Ошибка при закрытии соединения с БД: %v", err)
	}
}
