package handlers

import (
	"database/sql"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"net/http"
)

func DeleteTag(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
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
	if r.Method != http.MethodPost {
		logg.ErrorLog.Println("Trying open page with incorrect method")
		ErrorHandler(w, http.StatusMethodNotAllowed, "")
		return
	}
	tagId := r.FormValue("tag")
	err = DeleteTagEveryWhere(tagId, storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Error with deleting tags")
		return
	}
	logg.InfoLog.Printf("Tag with ID: %v has been deleted", tagId)
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func DeleteTagEveryWhere(tagId string, storage *sql.DB) error {
	query, err := storage.Prepare(`DELETE FROM tags WHERE id = ?`)
	if err != nil {
		return err
	}
	_, err = query.Exec(tagId)
	if err != nil {
		return err
	}
	query, err = storage.Prepare(`DELETE FROM posts_tags WHERE tag_id = ?`)
	if err != nil {
		return err
	}
	_, err = query.Exec(tagId)
	if err != nil {
		return err
	}
	return nil
}
