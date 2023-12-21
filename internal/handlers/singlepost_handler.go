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
)

func SinglePostHandler(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
	err := limiter.Limit()
	if err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, "Too many requests", http.StatusBadRequest)
		return
	}
	userId, isLogged := internal.UserIsLogged(r, storage)

	if r.Method != http.MethodGet {
		logg.ErrorLog.Println("Trying to open page with incorrect method")
		ErrorHandler(w, http.StatusMethodNotAllowed, "")
		return
	}
	if r.URL.Query().Get("id") != "" {
		var post entities.Post
		var data tmpl.Data
		postId, err := strconv.Atoi(r.URL.Query().Get("id"))
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusBadRequest, "")
			return
		}

		post, err = getPostById(postId, storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusNotFound, "No such post")
			return
		}

		post.Author, err = getUserNameById(post.UserId, storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}

		tagsIDs, err := getPostTagsByPostId(strconv.Itoa(post.Id), storage)
		if err != nil {
			logg.ErrorLog.Println("Cant fetch tag")
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}

		for _, v := range tagsIDs {
			tag, err := getTagByTagId(v, storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Database error")
				return
			}
			post.Tags = append(post.Tags, tag)
		}

		post.Comments, err = getCommentsByPostId(postId, storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}

		post.CommentsCount = len(post.Comments)

		post.ReactionsCount, err = getPostRate(post.Id, storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}
		if isLogged {
			data.Username, err = getUserNameById(userId, storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Database error")
				return
			}
			role, err := getRoleByUserId(userId, storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Database error")
				return
			}
			data.Role = role
		}
		data.IsLogged = isLogged
		data.Posts = post
		tmpl.RenderTemplate(w, "show_post_page.html", data)
	} else {
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
	}
}

func getPostById(postID int, storage *sql.DB) (entities.Post, error) {
	var post entities.Post
	var imgName sql.NullString
	row := storage.QueryRow("SELECT id, title, body, user_id, img_name FROM posts WHERE id = ?", postID)
	err := row.Scan(&post.Id, &post.Title, &post.Body, &post.UserId, &imgName)
	if err != nil {
		return post, err
	}
	if imgName.Valid {
		post.ImageName = imgName.String
	}
	return post, nil
}

func getCommentsByPostId(postID int, storage *sql.DB) ([]entities.Comment, error) {
	var comments []entities.Comment
	rows, err := storage.Query("SELECT id, post_id, user_id, body FROM comments WHERE post_id = ?", postID)
	if err != nil {
		return comments, err
	}
	defer rows.Close()
	for rows.Next() {
		comment := entities.Comment{}

		err := rows.Scan(&comment.Id, &comment.PostId, &comment.UserId, &comment.Body)
		if err != nil {
			return nil, err
		}
		comment.Author, err = getUserNameById(comment.UserId, storage)
		if err != nil {
			return nil, err
		}
		comment.ReactionsCount, err = getCommentRate(comment.Id, storage)
		comments = append(comments, comment)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}

func getUserNameById(userID string, storage *sql.DB) (string, error) {
	var username string
	row := storage.QueryRow("SELECT nickname FROM users WHERE id = ?", userID)
	err := row.Scan(&username)
	if err != nil {
		return "", err
	}
	return username, nil
}

func getTagByTagId(tagId string, storage *sql.DB) (string, error) {
	var tag string
	row := storage.QueryRow("SELECT name FROM tags WHERE id = ?", tagId)
	err := row.Scan(&tag)
	if err != nil {
		return tag, err
	}
	return tag, nil
}

func getPostTagsByPostId(postID string, storage *sql.DB) ([]string, error) {
	var tags []string
	rows, err := storage.Query("SELECT tag_id FROM posts_tags WHERE post_id = ?", postID)
	if err != nil {
		return tags, err
	}
	defer rows.Close()

	for rows.Next() {
		var tag string
		err := rows.Scan(&tag)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func getCommentCount(postID int, storage *sql.DB) (int, error) {
	var count int
	records, err := storage.Query("SELECT id FROM comments WHERE post_id = ?", postID)
	if err != nil {
		return count, err
	}
	defer records.Close()

	for records.Next() {
		count++
	}

	return count, nil
}
