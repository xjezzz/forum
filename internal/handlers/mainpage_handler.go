package handlers

import (
	"database/sql"
	"fmt"
	"forum-project/entities"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"forum-project/internal/tmpl"
	"html/template"
	"net/http"
	"strconv"
)

func MainPageHandler(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
	err := limiter.Limit()
	if err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, "Too many requests", http.StatusBadRequest)
		return
	}
	userId, isLogged := internal.UserIsLogged(r, storage)
	data := tmpl.Data{
		UserId:   userId,
		IsLogged: isLogged,
	}

	if isLogged {
		data.Username, err = getUserNameById(userId, storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}
		data.IsRequested, err = getIsRequested(userId, storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}
		data.Role, err = getRoleByUserId(userId, storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}
		if data.Role == "moderator" {
			data.Reports, err = getReportStatus(data.Username, storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Database error")
				return
			}
			err = deleteReports(data.Reports, storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Database error")
				return
			}
		}
		data.Actions, err = getAllActionsForUserId(userId, storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}
		for _, action := range data.Actions {
			err = closeActionNotify(action.Id, storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Error with actions from database")
				return
			}
		}
	}

	if r.URL.Query().Get("tag") != "" {
		post, err := getPostsByTag(r.URL.Query().Get("tag"), storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}

		data.Posts = post
	} else {
		posts, err := getAllPost(storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Error with getting posts from database")
			return
		}

		data.Posts = posts

	}
	data.Tags, err = getAllTags(storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Error with getting tags from database")
		return
	}

	tmpl, err := template.ParseFiles("web/templates/main_page.html")
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Database error")
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Database error")
		return
	}
}

func getAllPost(storage *sql.DB) ([]entities.Post, error) {
	records, err := storage.Query("SELECT id, user_id, title, body FROM posts ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer records.Close()

	var result []entities.Post

	for records.Next() {
		var post entities.Post

		err = records.Scan(&post.Id, &post.UserId, &post.Title, &post.Body)
		if err != nil {
			return nil, err
		}

		post.Author, err = getUserNameById(post.UserId, storage)
		if err != nil {
			return nil, err
		}
		tagsIDs, err := getPostTagsByPostId(strconv.Itoa(post.Id), storage)
		if err != nil {
			return nil, err
		}
		post.ReactionsCount, err = getPostRate(post.Id, storage)
		if err != nil {
			return nil, err
		}
		post.CommentsCount, err = getCommentCount(post.Id, storage)
		if err != nil {
			return nil, err
		}
		if len(tagsIDs) != 0 {
			for _, v := range tagsIDs {
				tag, err := getTagByTagId(v, storage)
				if err != nil {
					return nil, err
				}
				post.Tags = append(post.Tags, tag)
			}

			result = append(result, post)
		} else {

			tag, err := getTagByTagId("1", storage)
			if err != nil {
				return nil, err
			}
			post.Tags = append(post.Tags, tag)
			result = append(result, post)
		}
	}

	return result, nil
}

func getPostsByTag(tag string, storage *sql.DB) ([]entities.Post, error) {
	var result []entities.Post
	record := storage.QueryRow("SELECT id FROM tags WHERE name = ?", tag)
	var tagId string
	err := record.Scan(&tagId)
	if err != nil {
		return nil, err
	}
	records, err := storage.Query("SELECT post_id FROM posts_tags WHERE tag_id = ? ORDER BY post_id DESC", tagId)
	if err != nil {
		return result, err
	}

	var postsIds []int

	for records.Next() {
		var postId int
		err = records.Scan(&postId)
		if err != nil {
			return nil, err
		}
		postsIds = append(postsIds, postId)
	}

	for _, postID := range postsIds {
		record = storage.QueryRow("SELECT id, user_id, title, body FROM posts WHERE id = ? ORDER BY id DESC", postID)
		var post entities.Post

		err = record.Scan(&post.Id, &post.UserId, &post.Title, &post.Body)
		if err != nil {
			fmt.Println("lol")

			return nil, err
		}
		post.Author, err = getUserNameById(post.UserId, storage)
		if err != nil {
			return result, nil
		}
		post.ReactionsCount, err = getPostRate(post.Id, storage)
		if err != nil {
			return nil, err
		}
		post.CommentsCount, err = getCommentCount(post.Id, storage)
		if err != nil {
			return nil, err
		}

		tagsIDs, err := getPostTagsByPostId(strconv.Itoa(post.Id), storage)
		if err != nil {
			return result, err
		}

		for _, v := range tagsIDs {
			tag, err := getTagByTagId(v, storage)
			if err != nil {
				return result, err
			}
			post.Tags = append(post.Tags, tag)
		}

		result = append(result, post)
	}

	return result, nil
}

func getAllUserPosts(userId string, storage *sql.DB) ([]entities.Post, error) {
	records, err := storage.Query("SELECT id, title, body FROM posts WHERE user_id = ? ORDER BY id DESC", userId)
	if err != nil {
		return nil, err
	}
	defer records.Close()

	var result []entities.Post

	for records.Next() {
		var post entities.Post

		post.Author, err = getUserNameById(userId, storage)

		if err != nil {
			return nil, err
		}
		err = records.Scan(&post.Id, &post.Title, &post.Body)
		if err != nil {
			return nil, err
		}
		tagsIDs, err := getPostTagsByPostId(strconv.Itoa(post.Id), storage)
		if err != nil {
			return nil, err
		}
		post.ReactionsCount, err = getPostRate(post.Id, storage)
		if err != nil {
			return nil, err
		}

		post.CommentsCount, err = getCommentCount(post.Id, storage)
		if err != nil {
			return nil, err
		}

		for _, v := range tagsIDs {
			tag, err := getTagByTagId(v, storage)
			if err != nil {
				return result, err
			}
			post.Tags = append(post.Tags, tag)
		}

		result = append(result, post)
	}
	return result, nil
}

func getLikedOrDislikedPosts(action, userId string, storage *sql.DB) ([]entities.Post, error) {
	records, err := storage.Query("SELECT posts.id, posts.title, posts.body, posts.user_id FROM reactions, posts WHERE reactions.post_id = posts.id AND reactions.user_id = ? AND reactions.is_like = ? ORDER BY posts.id DESC", userId, action)
	if err != nil {
		return nil, err
	}
	var result []entities.Post
	for records.Next() {
		var id int
		var title, body, userId2 string

		username, err := getUserNameById(userId, storage)
		if err != nil {
			return result, nil
		}

		err = records.Scan(&id, &title, &body, &userId2)
		if err != nil {
			return nil, err
		}
		reactionsCount, err := getPostRate(id, storage)
		if err != nil {
			return nil, err
		}

		commentsCount, err := getCommentCount(id, storage)
		if err != nil {
			return nil, err
		}
		post := entities.Post{
			Id:             id,
			UserId:         userId2,
			Author:         username,
			Title:          title,
			Body:           body,
			ReactionsCount: reactionsCount,
			CommentsCount:  commentsCount,
		}

		tagsIDs, err := getPostTagsByPostId(strconv.Itoa(post.Id), storage)
		if err != nil {
			return nil, err
		}

		for _, v := range tagsIDs {
			tag, err := getTagByTagId(v, storage)
			if err != nil {
				return nil, err
			}
			post.Tags = append(post.Tags, tag)
		}

		result = append(result, post)
	}
	return result, nil
}

func getIsRequested(userId string, storage *sql.DB) (bool, error) {
	var isRequest bool
	row := storage.QueryRow("SELECT request FROM users WHERE id = ?", userId)
	err := row.Scan(&isRequest)
	if err != nil {
		return false, err
	}
	return isRequest, nil
}

func getReportStatus(username string, storage *sql.DB) ([]entities.Report, error) {
	var reports []entities.Report
	rows, err := storage.Query(`SELECT id, status, report FROM reports WHERE nickname = ? AND status IS NOT NULL`, username)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var status *bool
		var report entities.Report
		err := rows.Scan(&report.Id, &report.Status, &report.Reason)
		if err != nil && status != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	return reports, nil
}
