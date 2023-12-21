package handlers

import (
	"database/sql"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"net/http"
)

func UpdateUser(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
	err := limiter.Limit()
	if err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, "Too many requests", http.StatusBadRequest)
		return
	}
	_, isLogged := internal.UserIsLogged(r, storage)
	if !isLogged {
		logg.InfoLog.Println("Unauthorized users cannot enter this page")
		ErrorHandler(w, http.StatusUnauthorized, "Unauthorized users cannot use this page")
		return
	}
	id := r.URL.Query().Get("id")
	role, err := getRoleByUserId(id, storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Database error")
		return
	}
	if role == "user" {
		isRequested, err := getIsRequested(id, storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}
		if isRequested {
			err := updateUserRole(id, "moderator", storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Database error")
				return
			}
			logg.InfoLog.Printf("User with ID: %v promoted to moderator", id)

		}
	} else if role == "moderator" {
		err := updateUserRole(id, "user", storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}
		logg.InfoLog.Printf("Moderator with ID: %v demoted to user", id)
	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func updateUserRole(userId, role string, storage *sql.DB) error {
	query, err := storage.Prepare(`UPDATE users SET roles = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	_, err = query.Exec(role, userId)
	if err != nil {
		return err
	}
	query, err = storage.Prepare(`UPDATE users SET request = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	_, err = query.Exec(0, userId)
	if err != nil {
		return err
	}
	return nil
}
