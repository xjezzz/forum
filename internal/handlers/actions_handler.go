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

func Actions(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
	err := limiter.Limit()
	if err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, "Too many requests", http.StatusBadRequest)
		return
	}
	userId, isLogged := internal.UserIsLogged(r, storage)
	if !isLogged {
		logg.ErrorLog.Println("Unauthorized users do not have access to this page")
		ErrorHandler(w, http.StatusUnauthorized, "Unauthorized users do not have access to this page")
		return
	}
	if r.Method != http.MethodGet {
		logg.ErrorLog.Println("Trying open page with incorrect method")
		ErrorHandler(w, http.StatusMethodNotAllowed, "")
		return
	}
	var data tmpl.Data
	switch r.URL.Query().Get("show") {
	case "myposts":
		data.Posts, err = getAllUserPosts(userId, storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Error with getting posts from database")
			return
		}
	case "liked-posts":
		data.Posts, err = getLikedOrDislikedPosts("1", userId, storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}
	case "disliked-posts":
		data.Posts, err = getLikedOrDislikedPosts("0", userId, storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}
	case "mycomments":
		data.Comments, err = getCommentsByUserId(userId, storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}
		for i, v := range data.Comments {
			data.Comments[i].PostTitle, err = getPostTitleById(v.PostId, storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Database error")
				return
			}
		}
	}

	data.Username, err = getUserNameById(userId, storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Error with getting username from database")
		return
	}

	data.Actions, err = getAllActionsForUserId(userId, storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Error with getting actions from database")
		return
	}
	logg.InfoLog.Println("Activity page successfully loaded")
	tmpl.RenderTemplate(w, "action_page.html", data)
}

func closeActionNotify(actionId int, storage *sql.DB) error {
	query := "UPDATE actions SET is_read = ? WHERE id = ?"
	_, err := storage.Exec(query, "0", actionId)
	if err != nil {
		return err
	}
	return nil
}

func getAllActionsForUserId(userId string, storage *sql.DB) ([]entities.Actions, error) {
	var actions []entities.Actions
	records, err := storage.Query(`SELECT id, post_id, by_user_id, action FROM actions WHERE user_id = ? `, userId)
	if err != nil {
		return nil, err
	}
	for records.Next() {
		var byUser string
		var action entities.Actions
		err = records.Scan(&action.Id, &action.PostId, &byUser, &action.Action)

		if err := storage.QueryRow(`SELECT nickname FROM users WHERE id = ?`, byUser).Scan(&action.ByUser); err != nil {
			return nil, err
		}
		actions = append(actions, action)
	}

	return actions, nil
}

func getCommentsByUserId(userId string, storage *sql.DB) ([]entities.Comment, error) {
	records, err := storage.Query(`SELECT id, post_id, body FROM comments WHERE user_id = ?`, userId)
	if err != nil {
		return nil, err
	}
	var comments []entities.Comment
	for records.Next() {
		var comment entities.Comment
		err := records.Scan(&comment.Id, &comment.PostId, &comment.Body)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	return comments, err
}

func getPostTitleById(id int, storage *sql.DB) (string, error) {
	var title string
	record := storage.QueryRow(`SELECT title FROM posts WHERE id = ?`, id)
	err := record.Scan(&title)
	if err != nil {
		return "", err
	}

	return title, nil
}

func addActionToActions(postId, userId, action string, storage *sql.DB) error {
	var hostUser string
	if err := storage.QueryRow(`SELECT user_id FROM posts WHERE id = ?`, postId).Scan(&hostUser); err != nil {
		return err
	}
	query, err := storage.Prepare(`INSERT INTO actions(user_id, by_user_id, action, post_id, is_read) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	_, err = query.Exec(hostUser, userId, action, postId, "1")
	if err != nil {
		return err
	}

	return nil
}
