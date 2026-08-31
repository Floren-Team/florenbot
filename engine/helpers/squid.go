package helpers

import (
	gorm "gorm.io/gorm"
	"errors"
	"florenbot/engine/structs"
)

func CreateSquidRoom(db *gorm.DB, room structs.SquidRooms) error {
	newRoom := structs.SquidRooms{
		OwnerId: room.OwnerId,
		Members: room.Members,
	}

	if err := db.Create(&newRoom).Error; err != nil {
		return err
	}
	return nil
}

func JoinRoom(db *gorm.DB, roomId uint64, userId uint64) error {
	var room structs.SquidRooms
	if err := db.First(&room, roomId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("кімнату не знайдено")
		}
		return err
	}

	if room.Status != "open" {
		return errors.New("реєстрація до цієї кімнати вже закрита або гра вже почалася")
	}

	var existingMember structs.SquidMembers
	err := db.Where("room_id = ? AND user_id = ?", roomId, userId).First(&existingMember).Error
	if err == nil {
		return errors.New("ви вже приєдналися до цієї кімнати")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	member := structs.SquidMembers{
		RoomId: roomId,
		UserId: userId,
	}

	if err := db.Create(&member).Error; err != nil {
		return err
	}

	return nil
}

func GetRoomSquid(db *gorm.DB, owner_id uint64) (bool, error) {
    var room structs.SquidRooms
    err := db.Where("owner_id = ?", owner_id).First(&room).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return false, nil // Ошибка не в базе
        }
        return false, err // Ошибка 
    }
    return true, nil // Комната существует
}