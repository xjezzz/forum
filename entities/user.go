package entities

type User struct {
	Id          int
	Email       string
	PassHash    string
	Nickname    string
	Role        string
	IsRequested bool
}
