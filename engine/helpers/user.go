package helpers

import (
	"encoding/json"
	"errors"
	"florenbot/engine/cache"
	engine "florenbot/engine/mysql"
	"florenbot/engine/structs"
	"florenbot/helpers"
	"fmt"
	"os"
	"strconv"
	"time"

	"gorm.io/gorm"
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
            "username":            user.Username,
            "balance":             user.Balance,
            "promocode":           user.PromoCode,
            "clan_id":             func() uint64 { if user.ClanID != nil { return *user.ClanID }; return 0 }(),
            "negative_reputation": user.NegativeReputation,
            "positive_reputation": user.PositiveReputation,
            "losses":              user.Losses,
            "wins":                user.Wins,
            "euro":                user.Euro,
        }
        _ = cache.HMSet(key, vals)
        cache.Expire(key, 10*time.Minute)
    case "local":
        _ = os.MkdirAll("cache", 0755)
        filename := fmt.Sprintf("cache/user_%d.json", user.ID)
        fileData, _ := json.Marshal(user)
        _ = os.WriteFile(filename, fileData, 0644)
    }
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
        if err == nil && len(data) > 0 {
            user.ID = tgID
            user.Username = data["username"]
            
            // Парсинг числовых значений
            if val, err := strconv.ParseFloat(data["balance"], 64); err == nil { user.Balance = val }
            if val, err := strconv.ParseFloat(data["euro"], 64); err == nil { user.Euro = val }
            if val, err := strconv.ParseUint(data["clan_id"], 10, 64); err == nil && val != 0 { v := val; user.ClanID = &v }
            
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

    // 2. Запрос к базе данных
    // ВАЖНО: Удален Preload("Role"), так как роль теперь привязана к chat_members, а не к пользователю
    err := engine.DB.Where("id = ?", tgID).First(&user).Error
    
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return user, fmt.Errorf("user %d not found", tgID)
        }
        return user, err
    }

    // 3. Сохранение в кеш после успешного запроса
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
