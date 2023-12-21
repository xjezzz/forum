package handlers

import (
	"database/sql"
	"forum-project/internal"
	"net/http"
	"net/url"

	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
)

const (
	GoogleAuthURL      = "https://accounts.google.com/o/oauth2/auth"
	GoogleTokenURL     = "https://oauth2.googleapis.com/token"
	GoogleUserInfoURL  = "https://www.googleapis.com/oauth2/v1/userinfo?access_token="
	GoogleClientID     = "303815876793-hnt2ehi67kanbstk4u1h0586iqm8d0bd.apps.googleusercontent.com"
	GoogleClientSecret = "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"
	GoogleRedirectURL  = "https://localhost:8080/oauth2callback-google"
)

func HandleGoogleLogin(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
	err := limiter.Limit()
	if err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, "Too many requests", http.StatusBadRequest)
		return
	}
	_, isLogged := internal.UserIsLogged(r, storage)

	if isLogged {
		logg.InfoLog.Println("Already authorized user")
		http.Redirect(w, r, "/main", http.StatusSeeOther)
	}
	switch r.Method {
	case http.MethodGet:
		data := url.Values{}
		data.Set("client_id", GoogleClientID)
		data.Set("redirect_uri", GoogleRedirectURL)
		data.Set("response_type", "code")
		data.Set("scope", "openid profile email")
		authURL := GoogleAuthURL + "?" + data.Encode()
		http.Redirect(w, r, authURL, http.StatusFound)
	default:
		ErrorHandler(w, http.StatusMethodNotAllowed, "")
		return
	}
}
