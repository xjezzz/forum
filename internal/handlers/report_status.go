package handlers

import (
	"database/sql"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"net/http"
)

func ReportStatus(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
	err := limiter.Limit()
	if err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, "Too many requests", http.StatusBadRequest)
		return
	}
	userId, isLogged := internal.UserIsLogged(r, storage)
	if !isLogged {
		logg.InfoLog.Println("Unauthorized users cant make reactions")
		ErrorHandler(w, http.StatusUnauthorized, "Unauthorized users cant make reactions")
		return
	}
	if r.Method != http.MethodPost {
		logg.ErrorLog.Println("Trying to open page with incorrect method")
		ErrorHandler(w, http.StatusMethodNotAllowed, "")
		return
	}
	role, err := getRoleByUserId(userId, storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Error with getting role from database")
		return
	}
	if role != "admin" {
		logg.ErrorLog.Printf("User with ID: %v don't have access to this page", userId)
		ErrorHandler(w, http.StatusInternalServerError, "Page for administration only")
		return
	}
	status := r.FormValue("status")
	reportId := r.FormValue("id")

	switch status {
	case "accept":
		err := removePost(r.FormValue("post-id"), storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Error with deleting post from database")
			return
		}
		err = updateReportStatus("1", reportId, storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Error with deleting report from database")
			return
		}
		logg.InfoLog.Printf("Report with ID: %v has been accepted by admin. Post has been deleted", reportId)
	case "decline":
		err = updateReportStatus("0", reportId, storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Error with deleting report from database")
			return
		}
		logg.InfoLog.Printf("Report with ID: %v has been declimed by admin", reportId)

	}
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func updateReportStatus(status, reportId string, storage *sql.DB) error {
	query, err := storage.Prepare(`UPDATE reports SET status = $1 WHERE id = $2`)
	if err != nil {
		return err
	}
	_, err = query.Exec(status, reportId)
	if err != nil {
		return err
	}
	return nil
}
