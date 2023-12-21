package handlers

import (
	"database/sql"
	"forum-project/entities"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"forum-project/internal/tmpl"
	"net/http"
	"strings"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
	err := limiter.Limit()
	if err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, "Too many requests", http.StatusBadRequest)
		return
	}
	_, isLogged := internal.UserIsLogged(r, storage)
	if isLogged {
		logg.InfoLog.Println("Authorized users first must log out")
		http.Redirect(w, r, "/main", http.StatusSeeOther)
	}
	switch r.Method {
	case http.MethodGet:
		{
			tmpl.RenderTemplate(w, "registration_page.html", nil)
		}
	case http.MethodPost:
		{
			nick := r.FormValue("nick")
			if len(strings.Trim(nick, " ")) < 4 {
				logg.InfoLog.Println("Short username")
				ErrorHandler(w, http.StatusBadRequest, "Short username")
				return
			}

			email := strings.ToLower(r.FormValue("email"))
			pwd := r.FormValue("psw")

			err := createUser(&entities.User{
				Email:    email,
				PassHash: pwd,
				Nickname: nick,
			}, storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusBadRequest, "Username or email already taken")
				return
			}
			logg.InfoLog.Printf("New user with nickname '%v' has been registired", nick)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
		}
	default:
		{
			logg.ErrorLog.Println("Trying to open page with incorrect method")
			ErrorHandler(w, http.StatusMethodNotAllowed, "")
			return
		}
	}
}

func createUser(user *entities.User, storage *sql.DB) error {
	records := `INSERT INTO users(email, pass_hash, nickname, roles, request) VALUES (?, ?, ?, ?, ?);`
	query, err := storage.Prepare(records)
	if err != nil {
		return err
	}
	userId, err := query.Exec(user.Email, user.PassHash, user.Nickname, "user", false)
	if err != nil {
		return err
	}
	newUser, err := userId.LastInsertId()
	if err != nil {
		return err
	}
	logg.InfoLog.Printf("User with user Id %v save to storage", newUser)
	return nil
}
