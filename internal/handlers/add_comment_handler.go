package handlers

import (
	"database/sql"
	"forum-project/entities"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"net/http"
	"strconv"
)

func AddCommentHandler(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
	err := limiter.Limit()
	if err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, "Too many requests", http.StatusBadRequest)
		return
	}
	userId, isLogged := internal.UserIsLogged(r, storage)
	if !isLogged {
		logg.ErrorLog.Println("Unauthorized users cant write comments")
		ErrorHandler(w, http.StatusUnauthorized, "Unauthorized users cant write comments")
		return
	}
	if r.Method != http.MethodPost {
		logg.ErrorLog.Println("Trying open page with incorrect method")
		ErrorHandler(w, http.StatusMethodNotAllowed, "")
		return
	}

	commentBody := r.FormValue("comment")
	postId := r.FormValue("post-id")
	id, err := strconv.Atoi(postId)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "")
		return
	}

	comment := entities.Comment{
		PostId: id,
		UserId: userId,
		Body:   commentBody,
	}

	err = addComment(storage, comment)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Database error")
		return

	}
	err = addActionToActions(postId, userId, "commented", storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Database error")
		return
	}
	logg.InfoLog.Println("Comment was added")
	http.Redirect(w, r, "post?id="+postId, http.StatusSeeOther)
}

func addComment(db *sql.DB, comment entities.Comment) error {
	record := `INSERT INTO comments(user_id, post_id, user_id, body) VALUES (?, ?, ?, ?)`

	query, err := db.Prepare(record)
	if err != nil {
		return err
	}

	_, err = query.Exec(comment.UserId, comment.PostId, comment.UserId, comment.Body)
	if err != nil {
		return err
	}

	return nil
}
