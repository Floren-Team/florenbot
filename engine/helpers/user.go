package helpers

import (
	"encoding/json"
	"errors"
	"florenbot/engine/cache"
	engine "florenbot/engine/mysql"
	structs "florenbot/engine/structs"
	"florenbot/helpers"
	"fmt"
	"gorm.io/gorm"
	"os"
	"strconv"
	"time"
	"log"
)

func GetRole(userID uint64) (string, error) {
	var member structs.ClanMember
	err := engine.DB.Select("role").Where("user_id = ?", userID).First(&member).Error
	if err != nil {
		return "", err
	}
	return member.Role, nil
}

func IsUserBanned(userID uint64) bool {
	var count int64
	engine.DB.Model(&structs.Ban{}).Where("user_id = ?", userID).Count(&count)
	return count > 0
}

func GetUserIDByUsername(username string) uint64 {
	var user structs.User
	engine.DB.Select("id").Where("username = ?", username).First(&user)
	return user.ID
}

func GetUser(userID uint64) (*structs.User, error) {
	var user structs.User
	err := engine.DB.Select("first_name").Where("id = ?", userID).First(&user).Error
	return &user, err
}

// GetUserByID получает пользователя из кеша

func GetUserByIDDB(userID uint64) (structs.User, error) {
	var user structs.User
	err := engine.DB.Where("id = ?", userID).First(&user).Error
	return user, err
}

// saveUserToCache записывает данные пользователя в кеш (Redis или локальный файл)
func saveUserToCache(user structs.User) {
	cacheEngine := helpers.GetEnv("CACHE_ENGINE", "local")
	key := fmt.Sprintf("user:%d", user.ID)

	switch cacheEngine {
	case "redis":
		vals := map[string]interface{}{
			"username":  user.Username,
			"balance":   user.Balance,
			"promocode": user.PromoCode,
			"clan_id": func() uint64 {
				if user.ClanID != nil {
					return *user.ClanID
				}
				return 0
			}(),
			"negative_reputation": user.NegativeReputation,
			"positive_reputation": user.PositiveReputation,
			"losses":              user.Losses,
			"wins":                user.Wins,
			"euro":                user.Euro,
			"vip":                 user.Vip,
		    "first_name":				user.FirstName,
		}
		_ = cache.HMSet(key, vals)
		_ = cache.Expire(key, 10*time.Minute)
	case "local":
		_ = os.MkdirAll("cache", 0755)
		filename := fmt.Sprintf("cache/user_%d.json", user.ID)
		fileData, _ := json.Marshal(user)
		_ = os.WriteFile(filename, fileData, 0644)
	}
}

func IncrementMessageCount(userID uint64, username string, firstName string) {
    _, err := GetOrCreateUser(userID, username, firstName)
    if err != nil {
        log.Printf("Error getting or creating user: %v", err)
        return
    }

    log.Printf("Incrementing message count for user %d", userID)
    
    result := engine.DB.Model(&structs.User{}).
        Where("id = ?", userID).
        Update("message_count", gorm.Expr("COALESCE(message_count, 0) + 1"))

    if result.Error != nil {
        log.Printf("Error incrementing message count: %v", result.Error)
        return
    }

    if result.RowsAffected == 0 {
        log.Printf("User %d not found or no update performed", userID)
    }
}

func GetTopByBalance(limit int) ([]structs.User, error) {
	var topUsers []structs.User
	
	// Сортуємо за balance у спадному порядку
	err := engine.DB.Model(&structs.User{}).
		Order("balance DESC").
		Limit(limit).
		Find(&topUsers).Error
		
	return topUsers, err
}

func GetTopByMessages(limit int) ([]structs.User, error) {
    var topUsers []structs.User
    
    // Сортуємо за message_count у спадному порядку
    err := engine.DB.Model(&structs.User{}).
        Order("message_count DESC").
        Limit(limit).
        Find(&topUsers).Error
        
    return topUsers, err
}

// GetUserByID получает пользователя из кеша или БД
func GetUserByID(tgID uint64) (structs.User, error) {
	var user structs.User
	cacheEngine := helpers.GetEnv("CACHE_ENGINE", "local")
	key := fmt.Sprintf("user:%d", tgID)

	// 1. Попытка чтения из кеша
	switch cacheEngine {
	case "redis":
		data, err := cache.HGetAll(key)
		// Проверяем, что нет ошибки и что в кеше действительно есть данные (хотя бы одно поле)
		if err == nil && len(data) > 0 {
			user.ID = tgID
			user.Username = data["username"]
			user.FirstName = data["first_name"]

			// Безопасный парсинг чисел (используем вспомогательные функции для чистоты кода)
			if val, err := strconv.ParseFloat(data["balance"], 64); err == nil {
				user.Balance = val
			}
			if val, err := strconv.ParseFloat(data["euro"], 64); err == nil {
				user.Euro = val
			}
			
			// Парсинг ClanID (указатель)
			if val, err := strconv.ParseUint(data["clan_id"], 10, 64); err == nil && val != 0 {
				v := val
				user.ClanID = &v
			}

			// Парсинг VIP
			if val, err := strconv.ParseUint(data["vip"], 10, 64); err == nil {
				user.Vip = int(val)
			}

			// Парсинг остальных целочисленных полей
			user.NegativeReputation, _ = strconv.Atoi(data["negative_reputation"])
			user.PositiveReputation, _ = strconv.Atoi(data["positive_reputation"])
			user.Losses, _ = strconv.Atoi(data["losses"])
			user.Wins, _ = strconv.Atoi(data["wins"])
			user.PromoCode = data["promocode"]

			return user, nil
		}
	case "local":
		filename := fmt.Sprintf("cache/user_%d.json", tgID)
		if fileData, err := os.ReadFile(filename); err == nil {
			if err := json.Unmarshal(fileData, &user); err == nil {
				return user, nil
			}
		}
	}

	// 2. Запрос к базе данных, если в кеше нет или кеш отключен
	err := engine.DB.Where("id = ?", tgID).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return user, fmt.Errorf("пользователь %d не найден", tgID)
		}
		return user, err
	}

	// 3. Сохранение в кеш после успешного запроса из БД
	saveUserToCache(user)

	return user, nil
}

// GetOrCreateUser находит пользователя или создает нового, сохраняя его в БД и кеш
func GetOrCreateUser(tgID uint64, username string, firstName string) (structs.User, error) {
	user, err := GetUserByID(tgID)

	// Если пользователь найден, возвращаем его
	if err == nil {
		return user, nil
	}

	// Если ошибка не "не найден", возвращаем её
	if err.Error() != fmt.Sprintf("user %d not found", tgID) {
		return structs.User{}, err
	}

	// Создаем нового пользователя
	newUser := structs.User{
		ID:                 tgID,
		Username:           username,
		FirstName:          firstName,
		Balance:            300000,
		NegativeReputation: 0,
		PositiveReputation: 0,
	}

	// Сохраняем в БД
	if err := engine.DB.Create(&newUser).Error; err != nil {
		return structs.User{}, err
	}

	// Обязательно сохраняем в кеш сразу после создания
	saveUserToCache(newUser)

	return newUser, nil
}
func IsCommandRestricted(userID uint64, chatID uint64, command string) (bool, error) {
    var count int64
    err := engine.DB.Model(&structs.UserCommandRestriction{}).
        Where("user_id = ? AND chat_id = ? AND command = ?", userID, chatID, command).
        Count(&count).Error

    if err != nil {
        return false, err
    }
    
    return count > 0, nil
}


func GetLastBonusTime(userID uint64) (time.Time, error) {
    var user struct {
        LastBonusAt time.Time
    }

    err := engine.DB.Model(&structs.User{}).Select("last_bonus_at").Where("id = ?", userID).Scan(&user).Error

    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return time.Time{}, nil
        }
        return time.Time{}, err
    }

    return user.LastBonusAt, nil
}


func AddVip(userId uint64, vipLevel uint16) error {
	err := engine.DB.Model(&structs.User{}).Where("id = ?", userId).Update("vip", vipLevel).Error
	if err != nil {
		log.Printf("Ошибка при обновлении VIP уровня для пользователя %d: %v", userId, err)
		return errors.New("Ошибка при обновлении VIP уровня для пользователя")

	}


	activeAt := time.Now()
	result := activeAt.Add(time.Hour * 24 * 30)
	err = engine.DB.Exec("UPDATE users SET vip = ?, `vip_active_at` = ? WHERE id = ?", vipLevel, result, userId).Error
	if err != nil {
		log.Printf("Ошибка при обновлении VIP уровня для пользователя %d: %v", userId, err)
		return errors.New("Ошибка при обновлении VIP уровня для пользователя")
	}

	return nil
}

func UpdateLastBonusTime(userID uint64, t time.Time) error {
    // Используем Model(&User{}) для указания таблицы и Where для выбора конкретной строки.
    // Update обновляет только одно указанное поле.
    err := engine.DB.Model(&structs.User{}).Where("id = ?", userID).Update("last_bonus_at", t).Error
    
    if err != nil {
        log.Printf("Ошибка при обновлении времени последнего бонуса для пользователя %d: %v", userID, err)
        return err
    }
    
    return nil
}

// UpdateUserCache обновляет кеш пользователя, удаляя старые данные и записывая актуальные
func UpdateUserCache(user structs.User) {
	cacheEngine := helpers.GetEnv("CACHE_ENGINE", "local")
	key := fmt.Sprintf("user:%d", user.ID)

	switch cacheEngine {
	case "redis":
		// 1. Удаляем старый ключ, чтобы полностью очистить состояние
		_ = cache.Del(key)

		// 2. Снова записываем актуальные данные (используем ту же логику, что в saveUserToCache)
		saveUserToCache(user)

	case "local":
		// Для локального кеша просто перезаписываем файл (это эффективнее, чем удаление)
		filename := fmt.Sprintf("cache/user_%d.json", user.ID)
		_ = os.MkdirAll("cache", 0755)
		fileData, _ := json.Marshal(user)
		_ = os.WriteFile(filename, fileData, 0644)
	}
}

// func GetUserRole(userID uint64) (*structs.Role, error) {
// 	var user structs.User
// 	err := engine.DB.Preload("Role").First(&user, userID).Error
// 	return user.Role, err
// }

// AddUserToLosses инкрементирует поражения
func AddUserToLosses(userID uint64) error {
	return engine.DB.Model(&structs.User{}).Where("id = ?", userID).Update("losses", gorm.Expr("losses + 1")).Error
}

// AddUserToWins инкрементирует победы
func AddUserToWins(userID uint64) error {
	return engine.DB.Model(&structs.User{}).Where("id = ?", userID).Update("wins", gorm.Expr("wins + 1")).Error
}

// CreateUser создает пользователя
func CreateUser(tgID uint64, username, firstName string) (structs.User, error) {
	user := structs.User{
		ID:        tgID,
		Username:  username,
		FirstName: firstName,
	}
	if err := engine.DB.Create(&user).Error; err != nil {
		return structs.User{}, err
	}
	return user, nil
}

// GetUsers возвращает список всех пользователей (только ID)
func GetUsers() ([]structs.User, error) {
	var users []structs.User
	err := engine.DB.Select("id").Find(&users).Error
	return users, err
}
