package helpers

import "florenbot/engine/structs"

// IsUserAdmin проверяет права администратора для объекта роли
func IsUserAdmin(role *structs.Role) bool {
	// Если роль nil, прав нет
	if role == nil {
		return false
	}
	return role.BaseShort == "admin" || role.BaseShort == "owner" || role.BaseShort == "creator"
}

// IsUserOwnerOrCreator проверяет права владельца или создателя
func IsUserOwnerOrCreator(role *structs.Role) bool {
	if role == nil {
		return false
	}
	return role.BaseShort == "owner" || role.BaseShort == "creator"
}
