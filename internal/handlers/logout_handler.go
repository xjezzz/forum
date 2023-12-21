package handlers

import (
	"database/sql"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"net/http"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
	err := limiter.Limit()
	if err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, "Too many requests", http.StatusBadRequest)
		return
	}
	id, isLogged := internal.UserIsLogged(r, storage)
	if !isLogged {
		logg.ErrorLog.Println("Unauthorized users cant log out")
		http.Redirect(w, r, "/main", http.StatusSeeOther)
	}
	if r.Method != http.MethodGet {
		logg.ErrorLog.Println("Trying to open page with incorrect method")
		ErrorHandler(w, http.StatusMethodNotAllowed, "")
		return
	}

	cookie, err := r.Cookie("session_token")
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "")
		return
	}

	err = deleteSession(cookie.Value, storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Database error")
		return
	} else {
		logg.InfoLog.Printf("User with ID: %v successfully logged out", id)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func deleteSession(token string, storage *sql.DB) error {
	query, err := storage.Prepare("DELETE FROM sessions WHERE token = ?")
	if err != nil {
		return err
	}
	_, err = query.Exec(token)
	if err != nil {
		return err
	}
	return nil
}
