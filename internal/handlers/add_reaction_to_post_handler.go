package handlers

import (
	"database/sql"
	"forum-project/entities"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"net/http"
)

func AddReactionToPostHandler(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
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
	reaction := entities.Reaction{
		UserId: userId,
	}

	if r.FormValue("reaction") == "like" {
		reaction.IsLike = true
	} else if r.FormValue("reaction") == "dislike" {
		reaction.IsLike = false
	} else {
		logg.InfoLog.Println("Incorrect reaction")
		ErrorHandler(w, http.StatusBadRequest, "Only likes & dislikes")
		return
	}
	reaction.PostId = r.FormValue("post-id")

	err = createReactionToPost(&reaction, storage)
	if err != nil {
		logg.ErrorLog.Println(err)
		ErrorHandler(w, http.StatusInternalServerError, "Database error")
		return
	}
	if reaction.IsLike {
		err = addActionToActions(reaction.PostId, userId, "liked", storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}
	} else {
		err = addActionToActions(reaction.PostId, userId, "disliked", storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Database error")
			return
		}
	}
	logg.InfoLog.Println("Reaction to post pressed")

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}

func createReactionToPost(reaction *entities.Reaction, storage *sql.DB) error {
	var oldReaction entities.Reaction
	row := storage.QueryRow(`SELECT is_like, user_id FROM reactions WHERE post_id = ? AND user_id = ?`, reaction.PostId, reaction.UserId)
	err := row.Scan(&oldReaction.IsLike, &oldReaction.UserId)
	if err == nil {
		if oldReaction.IsLike != reaction.IsLike {
			query, err := storage.Prepare(`UPDATE reactions SET is_like = $1 WHERE post_id = $2 AND user_id = $3`)
			if err != nil {
				return err
			}
			_, err = query.Exec(reaction.IsLike, reaction.PostId, reaction.UserId)
			if err != nil {
				return err
			}
		}
	} else {
		records := `INSERT INTO reactions(is_like, user_id, post_id) VALUES (?, ?, ?);`
		query, err := storage.Prepare(records)
		if err != nil {
			return err
		}
		_, err = query.Exec(reaction.IsLike, reaction.UserId, reaction.PostId)
		if err != nil {
			return err
		}
	}

	return nil
}

func getPostRate(postID int, storage *sql.DB) (int, error) {
	var rate int
	records, err := storage.Query("SELECT is_like FROM reactions WHERE post_id = ?", postID)
	if err != nil {
		return rate, err
	}
	defer records.Close()

	for records.Next() {
		var isLike bool
		err = records.Scan(&isLike)
		if err != nil {
			return rate, err
		}
		if isLike {
			rate++
		} else {
			rate--
		}
	}
	return rate, nil
}
