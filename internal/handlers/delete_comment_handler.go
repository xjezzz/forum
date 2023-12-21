package handlers

import (
	"database/sql"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"net/http"
)

func DeleteComment(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
	if err := limiter.Limit(); err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, "Too many requests", http.StatusBadRequest)
		return
	}

	userId, isLogged := internal.UserIsLogged(r, storage)
	if !isLogged {
		logg.ErrorLog.Println("Unauthorized users cant delete comments")
		ErrorHandler(w, http.StatusUnauthorized, "Unauthorized users cant delete comments")
		return
	}

	role, err := getRoleByUserId(userId, storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Database error")
		return
	}

	switch r.Method {
	case http.MethodGet:
		{
			commentId := r.URL.Query().Get("id")

			authorOfComment, err := getAuthorOfComment(commentId, storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Database error")
				return
			}

			if role != "admin" && role != "moderator" && authorOfComment != userId {
				logg.ErrorLog.Println("Forbidden")
				ErrorHandler(w, http.StatusForbidden, "Forbidden")
				return
			}

			if commentId == "" {
				logg.ErrorLog.Println("Empty comment id")
				ErrorHandler(w, http.StatusBadRequest, "")
				return
			}
			if err = deleteComment(commentId, storage); err != nil {
				logg.ErrorLog.Println("StatusInternalServerError")
				ErrorHandler(w, http.StatusInternalServerError, "Database error")
				return
			}
			http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		}
	default:
		{
			logg.ErrorLog.Println("Trying to open page with incorrect method")
			ErrorHandler(w, http.StatusMethodNotAllowed, "")
			return
		}
	}
}

func deleteComment(commentId string, storage *sql.DB) error {
	_, err := storage.Exec(`DELETE FROM comments WHERE id = ?`, commentId)
	if err != nil {
		return err
	}
	_, err = storage.Exec(`DELETE FROM reactions WHERE comment_id = ?`, commentId)
	if err != nil {
		return err
	}

	logg.InfoLog.Printf("Comment with ID: %s deleted", commentId)
	return nil
}

func getAuthorOfComment(commentId string, storage *sql.DB) (string, error) {
	query := `SELECT user_id FROM comments WHERE id = ?`
	row := storage.QueryRow(query, commentId)
	var userId string
	if err := row.Scan(&userId); err != nil {
		return "", err
	}
	return userId, nil
}
