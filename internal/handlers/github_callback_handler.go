package handlers

import (
	"database/sql"
	"encoding/json"
	"forum-project/entities"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	logg "forum-project/internal/forum_logger"

	"github.com/google/uuid"
)

func HandleGithubCallback(clientID, clientSecret, tokenURL string, storage *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		data := url.Values{}
		data.Set("code", code)
		data.Set("client_id", GithubClientID)
		data.Set("client_secret", clientSecret)
		data.Set("redirect_uri", GithubRedirectURL)
		data.Set("grant_type", "authorization_code")
		resp, err := http.PostForm(GithubTokenURL, data)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Failed to exchange code for token")
			return
		}
		defer resp.Body.Close()
		var authBody auth

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logg.ErrorLog.Println(err)
		}
		values, _ := url.ParseQuery(string(body))
		accessToken := values.Get("access_token")
		if accessToken == "" {
			logg.ErrorLog.Println("Access token not found in response")
			ErrorHandler(w, http.StatusInternalServerError, "")
			return
		}

		err = readResponseByGithub(accessToken, &authBody)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Failed read response")
			return
		}

		userID, err := getAuthUserIdByName(authBody.Name, storage)
		if err != nil {
			err = createUser(&entities.User{
				PassHash: uuid.NewString(),
				Nickname: authBody.Name,
			}, storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Database error")
				return
			}
			userID, err = getAuthUserIdByName(authBody.Name, storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Database error")
				return
			}
		}
		token := uuid.NewString()

		expiredAt := time.Now().Add(300 * time.Second)
		session := entities.Session{
			UserId:  strconv.Itoa(userID),
			Token:   token,
			Expired: expiredAt,
		}
		cookie := &http.Cookie{
			Name:    "session_token",
			Value:   token,
			Expires: expiredAt,
		}

		http.SetCookie(w, cookie)

		err = createSession(session, storage)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "")
			return
		}

		logg.InfoLog.Printf("User with ID: %v successfully logged in", userID)

		http.Redirect(w, r, "/main", http.StatusSeeOther)
	}
}

func readResponseByGithub(token string, resp *auth) error {
	req, err := http.NewRequest(http.MethodGet, GithubUserInfoURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+token)
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	userInfo, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	resp.Name = parseUsername(string(userInfo))
	return nil
}

func parseUsername(body string) string {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(body), &result); err != nil {
		logg.ErrorLog.Println("Error parsing user information JSON:", err)
		return ""
	}
	username, ok := result["login"].(string)
	if !ok {
		logg.ErrorLog.Println("Error extracting username from JSON")
		return ""
	}
	return username
}

func getAuthUserIdByName(name string, storage *sql.DB) (int, error) {
	var id int
	record := storage.QueryRow("SELECT id FROM users WHERE nickname = ?", name)
	err := record.Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, err
}
