package engine

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

// InitDB инициализирует подключение к MySQL и создает таблицы
func InitDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		getEnv("DB_USER", "root"),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "3306"),
		getEnv("DB_NAME", "game_db"),
	)

	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("❌ Ошибка конфигурации MySQL: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("❌ Не удалось подключиться к MySQL (Ping): %v", err)
	}

	// Создание таблиц
	usersQuery := `CREATE TABLE IF NOT EXISTS users
	 (id BIGINT PRIMARY KEY,
	  username VARCHAR(255), 
	  balance FLOAT DEFAULT 1000,
	  promocode VARCHAR(32) UNIQUE NULL,
	  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP) 
	  ENGINE=InnoDB;`
	if _, err = DB.Exec(usersQuery); err != nil {
		log.Fatalf("❌ Ошибка создания таблицы users: %v", err)
	}

	blacklistQuery := `CREATE TABLE IF NOT EXISTS blacklists (id INT AUTO_INCREMENT PRIMARY KEY, user_id BIGINT NOT NULL, reason TEXT NOT NULL, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, CONSTRAINT fk_user FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE) ENGINE=InnoDB;`
	if _, err = DB.Exec(blacklistQuery); err != nil {
		log.Fatalf("❌ Ошибка создания таблицы blacklists: %v", err)
	}


	promocodeQuery := `
	CREATE TABLE IF NOT EXISTS promocodes (
		id INT AUTO_INCREMENT PRIMARY KEY,
		code VARCHAR(255) NOT NULL,
		amount INT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB;
	`
	if _, err = DB.Exec(promocodeQuery); err != nil {
		log.Fatalf("❌ Ошибка создания таблицы promocodes: %v", err)
	}
	
	log.Println("✅ Подключение к MySQL успешно")
}

// GetUserBalanceSQL получает баланс пользователя или создает его, если он новый
func GetUserBalanceSQL(id int64, username string) (int, error) {
	var balance int
	err := DB.QueryRow("SELECT balance FROM users WHERE id = ?", id).Scan(&balance)

	if err == sql.ErrNoRows {
		initialBalance := 1000
		_, err = DB.Exec("INSERT INTO users (id, username, balance) VALUES (?, ?, ?)", id, username, initialBalance)
		if err != nil {
			return 0, err
		}
		return initialBalance, nil
	}

	if err != nil {
		return 0, err
	}

	return balance, nil
}

func GetUser(id int64) (int64, error) {
	var telegram_id int64
	err := DB.QueryRow("SELECT id FROM users WHERE id = ?", id).Scan(&telegram_id)
	return telegram_id, err
}

func GetUserCode(userID int64) (string, error) {
    var code sql.NullString
    err := DB.QueryRow("SELECT promocode FROM users WHERE id = ?", userID).Scan(&code)
    
    if err != nil {
        return "", err
    }
    
    if !code.Valid {
        return "", nil 
    }
    return code.String, nil
}




func ActivateCode(id int64, code string) error {
	_, err := DB.Exec("UPDATE users SET promocode = ? WHERE id = ?", code, id)
	return err
}

func CreateCode(code string, amount int) error {
	_, err := DB.Exec("INSERT INTO promocodes (code, amount) VALUES (?, ?)", code, amount)
	return err
}

func GetCode(code string) (int, error) {
	var amount int
	err := DB.QueryRow("SELECT amount FROM promocodes WHERE code = ?", code).Scan(&amount)
	return amount, err
}

func DeleteCode(code string) error {
	// _, err := DB.Exec("DELETE FROM promocodes WHERE code = ?", code)
	// return err

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
    if err != nil {
        tx.Rollback()
        return err
    }

	return tx.Commit()

	
}

// UpdateBalanceSQL обновляет баланс пользователя
func UpdateBalanceSQL(id int64, amount int) error {
	_, err := DB.Exec("UPDATE users SET balance = balance + ? WHERE id = ?", amount, id)
	return err
}

// IsUserBanned проверяет наличие пользователя в черном списке
func IsUserBanned(id int64) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM blacklists WHERE user_id = ?)"
	err := DB.QueryRow(query, id).Scan(&exists)
	return exists, err
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}