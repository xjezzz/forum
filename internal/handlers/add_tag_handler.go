package handlers

import (
	"database/sql"
	"errors"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"net/http"
	"strings"
)

func CreateTag(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
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
	newTag := r.FormValue("tag")
	if len(newTag) < 3 || len(newTag) > 10 {
		logg.ErrorLog.Println("Incorrect new tag")
		ErrorHandler(w, http.StatusBadRequest, "Length of tag is not correct")
		return
	}
	newTag = strings.Title(newTag)
	err = addTag(newTag, storage)

	if err != nil {
		if err.Error() == "Tag already taken" {

			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusBadRequest, "Tag already has taken")
			return
		}
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Error with adding tag to database")
		return
	}
	logg.InfoLog.Printf("New tag named '%v' has been added", newTag)
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func addTag(tag string, storage *sql.DB) error {
	rows, err := storage.Query(`SELECT name from tags`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var tagName string
	for rows.Next() {
		err = rows.Scan(&tagName)
		if err != nil {
			return err
		}
		if tag == tagName {
			return errors.New("Tag already taken")
		}
	}

	query, err := storage.Prepare(`INSERT INTO tags(name) VALUES (?)`)
	if err != nil {
		return err
	}
	_, err = query.Exec(tag)
	if err != nil {
		return err
	}
	return nil
}
