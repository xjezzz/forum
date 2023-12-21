package handlers

import (
	"database/sql"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"net/http"
)

func CheckModerRequests(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
	err := limiter.Limit()
	if err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, "Too many requests", http.StatusBadRequest)
		return
	}
	userId, isLogged := internal.UserIsLogged(r, storage)
	if !isLogged {
		logg.InfoLog.Println("Unauthorized users cannot enter this page")
		ErrorHandler(w, http.StatusUnauthorized, "Unauthorized users cannot use this page")
		return
	}
	if r.Method != http.MethodGet {
		logg.InfoLog.Println("Trying to open page with incorrect method")
		ErrorHandler(w, http.StatusMethodNotAllowed, "")
		return
	}

	role, err := getRoleByUserId(userId, storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Database error")
		return
	}
	if role != "user" {
		logg.ErrorLog.Println("Forbidden")
		ErrorHandler(w, http.StatusForbidden, "Forbidden")
		return
	}
	err = addRequest(userId, storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Database error")
		return
	}
	logg.InfoLog.Printf("User with ID: %v send request for moderation", userId)
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func addRequest(userId string, storage *sql.DB) error {
	record := `UPDATE users SET request = ? WHERE id = ?`
	query, err := storage.Prepare(record)
	if err != nil {
		return err
	}
	_, err = query.Exec(true, userId)
	if err != nil {
		return err
	}
	return nil
}
