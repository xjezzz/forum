package entities

import "time"

type Session struct {
	UserId  string
	Token   string
	Expired time.Time
}
