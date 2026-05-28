package model

type Clans struct {
	Id          int64
	Name        string
	OwnerId     uint64
	InviteCode  string
	MemberCount int16
}
