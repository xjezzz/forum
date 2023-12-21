package entities

type Reaction struct {
	Id        int
	IsLike    bool
	UserId    string
	CommentId string
	PostId    string
}
