package handlers

import (
	"database/sql"
	"errors"
	"forum-project/entities"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"forum-project/internal/tmpl"
	"net/http"
	"strconv"
)

func EditPost(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
	if err := limiter.Limit(); err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, "Too many requests", http.StatusBadRequest)
		return
	}

	userId, isLogged := internal.UserIsLogged(r, storage)
	if !isLogged {
		logg.ErrorLog.Println("Unauthorized users cant edit posts")
		ErrorHandler(w, http.StatusUnauthorized, "Unauthorized users cant edit posts")
		return
	}
	switch r.Method {
	case http.MethodGet:
		{
			postId := r.URL.Query().Get("id")
			if postId == "" {
				logg.ErrorLog.Println("Invalid post id")
				ErrorHandler(w, http.StatusBadRequest, "Invalid post id")
				return
			}
			postID, err := strconv.Atoi(postId)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusBadRequest, "Invalid post id")
				return
			}
			post, err := getPostById(postID, storage)
			if userId != post.UserId {
				logg.ErrorLog.Println("You can edit only your posts")
				ErrorHandler(w, http.StatusForbidden, "You can edit only your posts")
				return
			}
			tmpl.RenderTemplate(w, "edit_post_page.html", post)
		}
	case http.MethodPost:
		{
			var post entities.Post
			postId := r.FormValue("id")
			post.Title = r.FormValue("title")
			post.Body = r.FormValue("post")

			file, handler, err := r.FormFile("myFile")
			if err != nil && !errors.Is(err, http.ErrMissingFile) {
				logg.InfoLog.Println("Size of file minimum 20 MB!")
				ErrorHandler(w, http.StatusBadRequest, "Invalid file")
				return
			}

			if err == nil {
				defer file.Close()
				post.ImageName, err = validImg(file, handler.Filename)
				if err != nil {
					logg.ErrorLog.Println(err)
					ErrorHandler(w, http.StatusBadRequest, "Invalid file")
					return
				}
			}
			if err := updatePost(postId, &post, storage); err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Database error")
				return
			}
			logg.InfoLog.Printf("Post with ID: %v has been edited", postId)
			http.Redirect(w, r, "/post?id="+postId, http.StatusSeeOther)
		}
	}
}

func updatePost(postId string, post *entities.Post, storage *sql.DB) error {
	query := "UPDATE posts SET title = ?, body = ?, img_name = ? WHERE id = ?"
	_, err := storage.Exec(query, post.Title, post.Body, post.ImageName, postId)
	if err != nil {
		return err
	}

	return nil
}
