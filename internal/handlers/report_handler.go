package handlers

import (
	"database/sql"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"net/http"
)

func ReportPost(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
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
	if r.Method != http.MethodPost {
		logg.ErrorLog.Println("Trying to open page with incorrect method")
		ErrorHandler(w, http.StatusMethodNotAllowed, "")
		return
	}
	name, err := getUserNameById(userId, storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Errorw with getting username from database")
		return
	}
	postId := r.FormValue("post-id")
	report := r.FormValue("report")
	err = addReport(postId, name, report, storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Error with adding report to database")
		return
	}
	logg.InfoLog.Printf("Report to post with ID: %v, was added by %v", postId, name)
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func addReport(postId, userId, report string, storage *sql.DB) error {
	query, err := storage.Prepare(`INSERT INTO reports(nickname, post_id, report) VALUES (?,?,?)`)
	if err != nil {
		return err
	}
	_, err = query.Exec(userId, postId, report)
	if err != nil {
		return err
	}
	return nil
}
