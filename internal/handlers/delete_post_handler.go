package handlers

import (
	"database/sql"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"net/http"
)

func DeletePost(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
	userId, isLogged := internal.UserIsLogged(r, storage)
	if !isLogged {
		logg.InfoLog.Println("Unauthorized users cannot enter this page")
		ErrorHandler(w, http.StatusUnauthorized, "Unauthorized users cannot use this page")
		return
	}
	if r.Method != http.MethodGet {
		logg.ErrorLog.Println("Trying open page with incorrect method")
		ErrorHandler(w, http.StatusMethodNotAllowed, "")
		return

	}
	id := r.URL.Query().Get("id")
	role, err := getRoleByUserId(userId, storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Error with getting role in database")
		return
	}
	if role != "moderator" && role != "admin" {
		logg.ErrorLog.Println("Users don't have access to this page")
		ErrorHandler(w, http.StatusForbidden, "Users don't have access to this page")
		return
	}
	err = removePost(id, storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Error with deleting post in database")
		return
	}
	logg.InfoLog.Printf("Post with ID: %v was deleted by user with ID: %v", id, userId)
	http.Redirect(w, r, "/main", http.StatusSeeOther)
}

func removePost(postId string, storage *sql.DB) error {
	_, err := storage.Exec(`DELETE FROM posts WHERE id = ?`, postId)
	if err != nil {
		return err
	}
	_, err = storage.Exec(`DELETE FROM comments WHERE post_id = ?`, postId)
	if err != nil {
		return err
	}
	_, err = storage.Exec(`DELETE FROM reactions WHERE post_id = ?`, postId)
	if err != nil {
		return err
	}
	_, err = storage.Exec(`DELETE FROM posts_tags WHERE post_id = ?`, postId)
	if err != nil {
		return err
	}
	return nil
}
