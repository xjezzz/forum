package handlers

import (
	"database/sql"
	"forum-project/entities"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"forum-project/internal/tmpl"
	"net/http"
	"strconv"
	"strings"
)

func EditComment(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
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
			commentId := r.URL.Query().Get("id")
			if commentId == "" {
				logg.ErrorLog.Println("Invalid comment id")
				ErrorHandler(w, http.StatusBadRequest, "Invalid comment id")
				return
			}
			comment, err := getCommentById(commentId, storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Database error")
			}
			if userId != comment.UserId {
				logg.ErrorLog.Println("You can edit only your comments")
				ErrorHandler(w, http.StatusForbidden, "You can edit only your comments")
				return
			}
			comment.Id, err = strconv.Atoi(commentId)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusBadRequest, "Invalid comment id")
				return
			}
			tmpl.RenderTemplate(w, "edit_comment_page.html", comment)
		}
	case http.MethodPost:
		{
			var comment entities.Comment
			commentId := r.FormValue("comment-id")
			comment.Body = r.FormValue("comment")
			postId := r.FormValue("post-id")
			if len(strings.TrimSpace(comment.Body)) < 10 {
				logg.InfoLog.Println("Tried to send too short comment")
				ErrorHandler(w, http.StatusBadRequest, "Too short comment")
				return
			}
			if err := updateComment(commentId, &comment, storage); err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Database error")
				return
			}
			logg.InfoLog.Printf("Comment with ID: %v has been edited", commentId)
			http.Redirect(w, r, "/post?id="+postId, http.StatusSeeOther)
		}
	default:
		logg.ErrorLog.Println("Trying to open page with incorrect method")
		ErrorHandler(w, http.StatusMethodNotAllowed, "")
		return
	}
}

func updateComment(commentId string, comment *entities.Comment, storage *sql.DB) error {
	query := "UPDATE comments SET body = ? WHERE id = ?"
	_, err := storage.Exec(query, comment.Body, commentId)
	if err != nil {
		return err
	}
	logg.InfoLog.Println("Comment updated")
	return nil
}

func getCommentById(id string, storage *sql.DB) (*entities.Comment, error) {
	var comment entities.Comment
	if err := storage.QueryRow("SELECT user_id, body, post_id FROM comments WHERE id = ?", id).Scan(&comment.UserId, &comment.Body, &comment.PostId); err != nil {
		return nil, err
	}
	return &comment, nil
}
