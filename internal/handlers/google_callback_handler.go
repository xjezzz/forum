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

type authCallback struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	IDToken     string `json:"id_token"`
}

type auth struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Locale        string `json:"locale"`
}

func HandleGoogleCallback(clientID, clientSecret, tokenURL string, storage *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")

		// Обмениваем код авторизации на токен доступа.
		data := url.Values{}
		data.Set("code", code)
		data.Set("client_id", GoogleClientID)
		data.Set("client_secret", GoogleClientSecret)
		data.Set("redirect_uri", GoogleRedirectURL)
		data.Set("grant_type", "authorization_code")

		resp, err := http.PostForm(GoogleTokenURL, data)
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
			ErrorHandler(w, http.StatusInternalServerError, "")
			return
		}
		err = readResponseByGoogle(body, &authBody)
		if err != nil {
			logg.ErrorLog.Println(err)
			ErrorHandler(w, http.StatusInternalServerError, "Failed read response")
			return
		}

		userID, err := getAuthUserIdByEmail(authBody.Email, storage)
		if err != nil {
			err = createUser(&entities.User{
				Email:    authBody.Email,
				PassHash: uuid.NewString(),
				Nickname: authBody.Name,
			}, storage)
			if err != nil {
				logg.ErrorLog.Println(err)
				ErrorHandler(w, http.StatusInternalServerError, "Database error")
				return
			}
			userID, err = getAuthUserIdByEmail(authBody.Email, storage)
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

func getAuthUserIdByEmail(email string, storage *sql.DB) (int, error) {
	var id int
	record := storage.QueryRow("SELECT id FROM users WHERE email = ?", email)
	err := record.Scan(&id)
	if err != nil {
		return -1, err
	}
	return id, err
}

func readResponseByGoogle(body []byte, resp *auth) error {
	var result authCallback

	err := json.Unmarshal(body, &result)
	if err != nil {
		return err
	}

	url := GoogleUserInfoURL + result.AccessToken

	response, err := http.Get(url)
	if err != nil {
		logg.ErrorLog.Println(err)
		return err
	}
	defer response.Body.Close()

	body2, err := io.ReadAll(response.Body)
	if err != nil {
		logg.ErrorLog.Println(err)
		return err
	}

	err = json.Unmarshal(body2, resp)
	if err != nil {
		return err
	}

	return nil
}
