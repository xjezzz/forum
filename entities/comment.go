package entities

type Comment struct {
	Id             int
	PostId         int
	UserId         string
	Body           string
	Author         string
	ReactionsCount int
	PostTitle      string
	Reactions      []Reaction
}
