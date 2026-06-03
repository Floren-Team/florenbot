package structs

import "database/sql"

type Report struct {
	AngryId sql.NullInt64 `json:"angry_id"`
	UserId  sql.NullInt64 `json:"user_id"`
	Text    string        `json:"text"`
}
