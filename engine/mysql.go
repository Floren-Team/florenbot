package engine

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// Функция для отладки (дебаггер)
func debug(format string, v ...interface{}) {
	if os.Getenv("DEBUG") == "true" {
		log.Printf("[DEBUG] "+format, v...)
	}
}

type UserPromo struct {
    Code   string
    Amount int
}

type Clans struct {
	Id   int64
	Name string
	OwnerId uint64
}

func InitDB() {
	var err error
	dbMode := getEnv("DB_MODE", "mysql")
	debug("Режим БД: %s", dbMode)

	if dbMode == "mysql" {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
			getEnv("DB_USER", "root"),
			getEnv("DB_PASSWORD", ""),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "3306"),
			getEnv("DB_NAME", "game_db"),
		)
		DB, err = sql.Open("mysql", dsn)
	} else {
		log.Print("Использую локальную БД (SQLite)...")

		if _, err := os.Stat("db"); os.IsNotExist(err) {
            log.Println("Директория 'db' не найдена, создаю её...")
            os.Mkdir("db", 0755)
        }
		DB, err = sql.Open("sqlite3", "bot.db")

		_, err = DB.Exec("PRAGMA journal_mode=WAL;")
        if err != nil {
            log.Printf("⚠️ Не удалось включить WAL-режим: %v", err)
        } else {
			log.Println("✅ WAL-режим включен")
		}
	}

	if err != nil {
		log.Fatalf("❌ Ошибка при открытии БД: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("❌ Не удалось подключиться к БД: %v", err)
	}

	// SQL-запросы для создания таблиц
	var queries []string
    if dbMode == "mysql" {
		queries = []string{
			// 1. Спочатку створюємо незалежні таблиці
			`CREATE TABLE IF NOT EXISTS clans (
				id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, 
				name VARCHAR(255) NOT NULL, 
				owner_id BIGINT NOT NULL, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			) ENGINE=InnoDB;`,

			`CREATE TABLE IF NOT EXISTS promocodes (
				id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, 
				code VARCHAR(255) NOT NULL UNIQUE, 
				amount INT NOT NULL, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			) ENGINE=InnoDB;`,

			// 2. Таблиця users з FOREIGN KEY
			`CREATE TABLE IF NOT EXISTS users (
				id BIGINT PRIMARY KEY, 
				username VARCHAR(255), 
				balance FLOAT DEFAULT 1000, 
				promocode VARCHAR(32), 
				clan_id INT DEFAULT NULL, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (clan_id) REFERENCES clans(id) ON DELETE SET NULL,
				FOREIGN KEY (promocode) REFERENCES promocodes(code) ON DELETE SET NULL
			) ENGINE=InnoDB;`,

			// 3. Таблиця, що залежить від users
			`CREATE TABLE IF NOT EXISTS blacklists (
				id INT NOT NULL AUTO_INCREMENT PRIMARY KEY, 
				user_id BIGINT NOT NULL, 
				reason TEXT NOT NULL, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
				CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
			) ENGINE=InnoDB;`,
		}
	} else {
		// SQLite (порядок створення критично важливий)
		queries = []string{
			`CREATE TABLE IF NOT EXISTS clans (
				id INTEGER PRIMARY KEY AUTOINCREMENT, 
				name VARCHAR(255) NOT NULL, 
				owner_id BIGINT NOT NULL, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,

			`CREATE TABLE IF NOT EXISTS promocodes (
				id INTEGER PRIMARY KEY AUTOINCREMENT, 
				code VARCHAR(255) NOT NULL UNIQUE, 
				amount INT NOT NULL, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,

			`CREATE TABLE IF NOT EXISTS users (
				id BIGINT PRIMARY KEY, 
				username VARCHAR(255), 
				balance FLOAT DEFAULT 1000, 
				promocode VARCHAR(32), 
				clan_id INTEGER, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY(clan_id) REFERENCES clans(id) ON DELETE SET NULL,
				FOREIGN KEY(promocode) REFERENCES promocodes(code) ON DELETE SET NULL
			);`,

			`CREATE TABLE IF NOT EXISTS blacklists (
				id INTEGER PRIMARY KEY AUTOINCREMENT, 
				user_id BIGINT NOT NULL, 
				reason TEXT NOT NULL, 
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, 
				FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
			);`,
		}
	}
	for _, query := range queries {
		debug("Выполнение SQL: %s", query)
		if _, err = DB.Exec(query); err != nil {
			log.Fatalf("❌ Ошибка создания таблицы: %v", err)
		}
	}

	log.Println("✅ Подключение к БД успешно")
}

func GetUserBalanceSQL(id int64, username string) (int, error) {
	debug("GetUserBalanceSQL: id=%d", id)
	var balance int
	err := DB.QueryRow("SELECT balance FROM users WHERE id = ?", id).Scan(&balance)

	if err == sql.ErrNoRows {
		debug("Пользователь не найден, создаем: %d", id)
		_, err = DB.Exec("INSERT INTO users (id, username, balance) VALUES (?, ?, ?)", id, username, 1000)
		return 1000, err
	}
	return balance, err
}

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

	tx.Commit()
	return err
}

func GetClanByOwnerID(owner_id uint64) (uint64, error) {
	var id uint64
	err := DB.QueryRow("SELECT id FROM clans WHERE owner_id = ?", owner_id).Scan(&id)
	return id, err
}

func GetClan(id uint64) (*Clans, error) {
	clan := &Clans{}
	query := "SELECT id, name, owner_id FROM clans WHERE id = ?"
	err := DB.QueryRow(query, id).Scan(&clan.Id, &clan.Name, &clan.OwnerId)
	return clan, err


}

func DeleteClan(id uint64) error { // Змінив uint64 на int64 (для SQL це зазвичай краще)
    tx, err := DB.Begin()
    if err != nil {
        return err
    }

    // 1. Спочатку обнуляємо clan_id у користувачів (щоб уникнути порушення FOREIGN KEY)
    _, err = tx.Exec("UPDATE users SET clan_id = NULL WHERE clan_id = ?", id)
    if err != nil {
        tx.Rollback()
        return err
    }

    // 2. Потім видаляємо сам клан
    _, err = tx.Exec("DELETE FROM clans WHERE id = ?", id)
    if err != nil {
        tx.Rollback()
        return err
    }

    return tx.Commit()
}

func ActivateCode(id int64, code string) error {
	debug("Активация кода: %s для пользователя %d", code, id)
	_, err := DB.Exec("UPDATE users SET promocode = ? WHERE id = ?", code, id)
	return err
}

func DeleteCode(code string) error {
	debug("Удаление кода: %s", code)
	tx, err := DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM promocodes WHERE code = ?", code)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("UPDATE users SET promocode = NULL WHERE promocode = ?", code)
	debug("Результат обнуления промокода у пользователей: err=%v", err)
	
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}



func GetPromocodesUser(id int64) ([]UserPromo, error) {
	code, err := GetUserCode(id)
    if err != nil || code == "" {
        return nil, err
    }

    amount, _ := GetCode(code)
    
    // Возвращаем список из одного элемента
    return []UserPromo{{Code: code, Amount: amount}}, nil
}

func GetCode(code string) (int, error) {
	debug("Проверка кода: %s", code)
	var amount int
	err := DB.QueryRow("SELECT amount FROM promocodes WHERE code = ?", code).Scan(&amount)
	return amount, err
}

// Вспомогательные функции...
func GetUser(id int64) (int64, error) {
	var telegram_id int64
	err := DB.QueryRow("SELECT id FROM users WHERE id = ?", id).Scan(&telegram_id)
	return telegram_id, err
}

func GetUserCode(userID int64) (string, error) {
	var code sql.NullString
	err := DB.QueryRow("SELECT promocode FROM users WHERE id = ?", userID).Scan(&code)
	if err != nil { return "", err }
	return code.String, nil
}

func CreateCode(code string, amount int) error {
	debug("Создание кода: %s, сумма: %d", code, amount)
	_, err := DB.Exec("INSERT INTO promocodes (code, amount) VALUES (?, ?)", code, amount)
	return err
}

func UpdateBalanceSQL(id int64, amount int) error {
	_, err := DB.Exec("UPDATE users SET balance = balance + ? WHERE id = ?", amount, id)
	return err
}

func IsUserBanned(id int64) (bool, error) {
	var exists bool
	err := DB.QueryRow("SELECT EXISTS(SELECT 1 FROM blacklists WHERE user_id = ?)", id).Scan(&exists)
	return exists, err
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists { return value }
	return defaultValue
}


func CloseDB() {
    if DB != nil {
        log.Println("💾 Закрытие подключения к БД...")
        DB.Close()
    }
}