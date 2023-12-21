package handlers

import (
	"database/sql"
	"forum-project/entities"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"forum-project/internal/tmpl"
	"net/http"
)

func Admin(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
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
	switch r.Method {
	case http.MethodGet:
		role, err := getRoleByUserId(userId, storage)
		if err != nil || role != "admin" {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusBadRequest, "do not have right access")
			return
		}
		users, err := getAllUser(storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}
		switch r.URL.Query().Get("show") {
		case "reports":

			reports, err := getReports(storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Error with getting reports from database")
				return
			}
			logg.InfoLog.Println("Reports page successfully loaded")
			tmpl.RenderTemplate(w, "reports_page.html", reports)
			return
		case "tags":
			TagsHandler(w, r, storage, limiter)
			logg.InfoLog.Println("Tags page successfully loaded")

			return
		}
		logg.InfoLog.Println("Page of users successfully loaded")

		tmpl.RenderTemplate(w, "admin_page.html", users)
	default:
		{
			logg.ErrorLog.Println("Trying to open page with incorrect method")
			ErrorHandler(w, http.StatusMethodNotAllowed, "")
			return
		}
	}
}

func getAllUser(storage *sql.DB) ([]entities.User, error) {
	records, err := storage.Query("SELECT id, nickname, roles, request FROM users")
	if err != nil {
		return nil, err
	}
	defer records.Close()

	var result []entities.User

	for records.Next() {
		var user entities.User

		err = records.Scan(&user.Id, &user.Nickname, &user.Role, &user.IsRequested)
		if err != nil {
			return nil, err
		}

		result = append(result, user)
	}

	return result, nil
}

func getRoleByUserId(userId string, storage *sql.DB) (string, error) {
	var role string
	row := storage.QueryRow("SELECT roles FROM users WHERE id = ?", userId)
	err := row.Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}

func getReports(storage *sql.DB) ([]entities.Report, error) {
	var reports []entities.Report

	rows, err := storage.Query("SELECT id, post_id, nickname, report FROM reports WHERE status IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		report := entities.Report{}

		err := rows.Scan(&report.Id, &report.PostId, &report.Nickname, &report.Reason)
		if err != nil {
			return nil, err
		}

		reports = append(reports, report)

	}

	return reports, nil
}

func deleteReports(reports []entities.Report, storage *sql.DB) error {
	for _, report := range reports {
		query, err := storage.Prepare(`DELETE FROM reports WHERE id = ?`)
		if err != nil {
			return err
		}
		_, err = query.Exec(report.Id)
		if err != nil {
			return err
		}
	}
	return nil
}
