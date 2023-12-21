package internal

import (
	"database/sql"
	logg "forum-project/internal/forum_logger"
	"net/http"
	"time"
)

// UserIsLogged function returns true and userId, if request has alive session
func UserIsLogged(r *http.Request, storage *sql.DB) (string, bool) {
	cookieToken, err := r.Cookie("session_token")
	if err != nil {
		return "", false
	}

	var expired time.Time
	var userID string

	record := storage.QueryRow("SELECT user_id, expired FROM sessions WHERE token = ?", cookieToken.Value)
	err = record.Scan(&userID, &expired)
	if err != nil {
		return "", false
	}

	if time.Now().After(expired) {
		return "", false
	}

	_, err = storage.Exec("UPDATE sessions SET expired = ? WHERE token = ?", time.Now().Add(300*time.Second), cookieToken.Value)
	if err != nil {
		logg.ErrorLog.Println(err)
		return "", false
	}

	return userID, true
}
