package helpers

import (
	"florenbot/engine/mysql"
	"florenbot/engine/structs"
	"log"
)

func IsUserVIP(user_id uint64) (bool, error) {
    var count int64
    
    err := mysql.DB.Model(&structs.User{}).Where("id = ? AND vip > 0", user_id).Count(&count).Error
    
    if err != nil {
        log.Printf("Помилка при перевірці VIP рівня користувача %d: %v", user_id, err)
        return false, err
    }
    
    return count > 0, nil
}