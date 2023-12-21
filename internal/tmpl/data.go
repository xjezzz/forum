package tmpl

import "forum-project/entities"

// Data struct for execute template
type Data struct {
	UserId      string
	IsLogged    bool
	Posts       interface{}
	Username    string
	Role        string
	IsRequested bool
	Tags        []entities.Tag
	Reports     []entities.Report
	Comments    []entities.Comment
	Actions     []entities.Actions
}

type AdminData struct {
	Username string
	Users    interface{}
}
