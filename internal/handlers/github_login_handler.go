package handlers

import (
	"database/sql"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	rl "forum-project/internal/request_limiter"
	"net/http"
	"net/url"
)

const (
	GithubAuthURL      = "https://github.com/login/oauth/authorize"
	GithubClientID     = "693cfe3209f1f8f9bf71"
	GithubClientSecret = "b3329f77cd2bb993394f1e0ecf59acc0a36b68f6"
	GithubRedirectURL  = "https://localhost:8080/oauth2callback-github"
	GithubTokenURL     = "https://github.com/login/oauth/access_token"
	GithubUserInfoURL  = "https://api.github.com/user"
)

func HandleGithubLogin(w http.ResponseWriter, r *http.Request, storage *sql.DB, limiter *rl.RateLimiter) {
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
		data.Set("client_id", GithubClientID)
		data.Set("redirect_uri", GithubRedirectURL)
		data.Set("response_type", "code")
		data.Set("scope", "profile email")
		data.Set("state", "random-state")
		authURL := GithubAuthURL + "?" + data.Encode()
		http.Redirect(w, r, authURL, http.StatusFound)
	default:
		ErrorHandler(w, http.StatusMethodNotAllowed, "")
		return
	}
}
