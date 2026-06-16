package helpers

import (
	engine "florenbot/engine"
	"florenbot/engine/structs"
	"log"
)

func CreateChat(chat structs.Chat) error {
	_, err := engine.DB.Exec("INSERT INTO chat (id, name, user_id) VALUES (?, ?, ?)", chat.Id, chat.Name, chat.UserId)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func GetChatById(id int64) structs.Chat {
	var chat structs.Chat
	err := engine.DB.QueryRow("SELECT id, name, user_id FROM chat WHERE id = ?", id).Scan(&chat.Id, &chat.Name, &chat.UserId)
	if err != nil {
		log.Println(err)
	}
	return chat
}

func GetChats() []structs.Chat {
	var chats []structs.Chat
	rows, err := engine.DB.Query("SELECT id, name, user_id FROM chat")
	if err != nil {
		log.Println(err)
	}
	for rows.Next() {
		var chat structs.Chat
		err := rows.Scan(&chat.Id, &chat.Name, &chat.UserId)
		if err != nil {
			log.Println(err)
		}
		chats = append(chats, chat)
	}
	return chats
}


func DeleteChat(id int64) error {
	_, err := engine.DB.Exec("DELETE FROM chat WHERE id = ?", id)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}