package helpers

import (
	"florenbot/engine/cache"
	engine "florenbot/engine/mysql"
	"florenbot/engine/structs"
	"florenbot/helpers"
	"fmt"
	"os"
	"errors"
	"time"
	"strconv"

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

func GetUser(userID uint64) (*structs.User, error) {
    var user structs.User
    err := engine.DB.Select("first_name").Where("id = ?", userID).First(&user).Error
    return &user, err
}

// GetUserByID получает пользователя из кеша или БД
func GetUserByID(tgID uint64) (structs.User, error) {
    var user structs.User
    key := fmt.Sprintf("user:%d", tgID)

    // 1. Попытка чтения из Redis
    if helpers.GetEnv("CACHE_ENGINE", "local") == "redis" {
        data, err := cache.HGetAll(key)
        if err == nil && len(data) > 0 {
            // Восстанавливаем данные из мапы в структуру
            user.ID = tgID
            user.Username = data["username"]
            if val, err := strconv.ParseFloat(data["balance"], 64); err == nil { user.Balance = val }
            if val, err := strconv.ParseFloat(data["euro"], 64); err == nil { user.Euro = val }
            
            if val, err := strconv.ParseUint(data["clan_id"], 10, 64); err == nil {
                v := int64(val)
                user.ClanID = &v
            }

			if val, err := strconv.ParseInt(data["negative_reputation"], 10, 64); err == nil {
				user.NegativeReputation = int(val)
			}

			if val, err := strconv.ParseInt(data["positive_reputation"], 10, 64); err == nil {
				user.PositiveReputation = int(val)
			}

			if val, err := strconv.ParseInt(data["losses"], 10, 64); err == nil {
				user.Losses = int(val)
			}

			if val, err := strconv.ParseInt(data["wins"], 10, 64); err == nil {
				user.Wins = int(val)
			}

			user.PromoCode = data["promocode"]
			user.Role = data["role"]
            
            return user, nil
        }
    }

    // 2. Чтение из БД через GORM
    err := engine.DB.First(&user, tgID).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return structs.User{}, nil 
        }
        return structs.User{}, err
    }

    // 3. Синхронизация кеша (если пользователь найден в БД)
    if helpers.GetEnv("CACHE_ENGINE", "local") == "redis" {
        vals := map[string]interface{}{
            "username":            user.Username,
            "balance":             user.Balance,
            "promocode":           user.PromoCode,
            "clan_id":             user.ClanID,
            "role":                user.Role,
            "negative_reputation": user.NegativeReputation,
            "positive_reputation": user.PositiveReputation,
            "losses":              user.Losses,
            "wins":                user.Wins,
            "euro":                user.Euro,
        }
        _ = cache.HMSet(key, vals)
        cache.Expire(key, 10*time.Minute)
    } else {
        // Локальное файловое кеширование
        _ = os.MkdirAll("cache", 0755)
        filename := fmt.Sprintf("cache/user_%d.json", user.ID)
        _ = os.WriteFile(filename, []byte(fmt.Sprintf("%+v", user)), 0644)
    }

    return user, nil
}

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
