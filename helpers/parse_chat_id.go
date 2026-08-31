package helpers

func ParseChatID(chat_id uint64) int64 {
	if chat_id > 0x8000000000000000 {
		return -int64(chat_id)
	}
	return int64(chat_id)
}
