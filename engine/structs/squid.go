package structs

type SquidRooms struct {
	ID  	 uint64  `gorm:"primaryKey;autoIncrement" json:"id"`
	OwnerId  uint64  `gorm:"not null;uniqueIndex" json:"owner_id"`
	Status   string `gorm:"default:'open'" json:"status"`

    Owner   User           `gorm:"foreignKey:OwnerId;references:ID;constraint:OnDelete:CASCADE;" json:"owner,omitempty"`

    Members []SquidMembers `gorm:"foreignKey:RoomId;references:ID;constraint:OnDelete:CASCADE;" json:"members,omitempty"`
}

type SquidMembers struct {
	UserId uint64 `gorm:"primaryKey;type:bigint unsigned" json:"user_id"`
    RoomId uint64 `gorm:"primaryKey;type:bigint unsigned" json:"room_id"`

	User   User   `gorm:"foreignKey:UserId;references:ID;constraint:OnDelete:CASCADE;" json:"user,omitempty"`

}