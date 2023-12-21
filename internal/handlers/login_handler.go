package handlers

import (
	"database/sql"
	"errors"
	"forum-project/entities"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"forum-project/internal/tmpl"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func LoginHandler(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
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
			tmpl.RenderTemplate(w, "login_page.html", nil)
		}
	case http.MethodPost:
		{
			email := strings.ToLower(r.FormValue("email"))
			pwd := r.FormValue("psw")
			id, err := getUserByEmail(email, pwd, storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusBadRequest, "Wrong email or password")
				return
			}

			token := uuid.NewString()
			expiredAt := time.Now().Add(300 * time.Second)
			session := entities.Session{
				UserId:  strconv.Itoa(id),
				Token:   token,
				Expired: expiredAt,
			}
			cookie := &http.Cookie{
				Name:    "session_token",
				Value:   token,
				Expires: expiredAt,
			}

			http.SetCookie(w, cookie)

			err = createSession(session, storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "")
				return
			}
			logg.InfoLog.Printf("User with ID: %v successfully logged in", id)

			http.Redirect(w, r, "/main", http.StatusSeeOther)

		}
	default:
		{
			logg.ErrorLog.Println("Trying to open page with incorrect method")
			ErrorHandler(w, http.StatusMethodNotAllowed, "")
			return
		}
	}
}

func getUserByEmail(email, password string, storage *sql.DB) (int, error) {
	var passHash string
	var id int
	record := storage.QueryRow("SELECT pass_hash, id FROM users WHERE email = ?", email)
	err := record.Scan(&passHash, &id)
	if err != nil {
		return -1, err
	}
	if password != passHash {
		return -1, errors.New("password do not match")
	}
	return id, err
}

func createSession(session entities.Session, storage *sql.DB) error {
	_, err := storage.Exec("DELETE FROM sessions WHERE token = ?", session.Token)
	if err == nil {
		logg.InfoLog.Println("Previous session has been deleted")
	}

	record := `INSERT INTO sessions(user_id, token, expired) VALUES (?, ?, ?);`
	query, err := storage.Prepare(record)
	if err != nil {
		return err
	}
	_, err = query.Exec(session.UserId, session.Token, session.Expired)
	if err != nil {
		return err
	}
	logg.InfoLog.Println("Session was written to storage")
	return nil
}
