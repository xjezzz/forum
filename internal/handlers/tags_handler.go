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

func TagsHandler(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
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
	role, err := getRoleByUserId(userId, storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Error with getting role from database")
		return
	}
	if role != "admin" {
		logg.ErrorLog.Println("Only administration have access to this page")
		ErrorHandler(w, http.StatusForbidden, "Only administration have access to this page")
		return
	}
	if r.Method != http.MethodGet {
		logg.ErrorLog.Println("Trying to open page with incorrect method")
		ErrorHandler(w, http.StatusMethodNotAllowed, "")
		return
	}

	tags, err := getAllTags(storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Error with getting tags from database")
		return
	}
	tmpl.RenderTemplate(w, "tags_page.html", tags)
}

func getAllTags(storage *sql.DB) ([]entities.Tag, error) {
	var tags []entities.Tag
	rows, err := storage.Query(`SELECT id, name FROM tags WHERE id != 1`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var tag entities.Tag
		err := rows.Scan(&tag.Id, &tag.Name)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}
