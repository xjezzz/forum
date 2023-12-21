package entities

type Post struct {
	Id             int
	Title          string
	Body           string
	CommentsCount  int
	ReactionsCount int
	Author         string
	UserId         string
	ImageName      string
	Tags           []string
	Comments       []Comment
}
