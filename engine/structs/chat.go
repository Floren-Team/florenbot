package structs

type Chat struct {
	Id   int64 `json:"id"`
	Name string `json:"name"`
	UserId int64 `json:"user_id"`
}