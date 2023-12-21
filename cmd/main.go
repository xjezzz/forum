package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"forum-project/config"
	"forum-project/internal"
	logg "forum-project/internal/forum_logger"
	"forum-project/internal/handlers"
	rl "forum-project/internal/request_limiter"
	"log"
	"net/http"
	"time"
)

// initLoggers function for initialize custom loggers
func main() {
	configuration, err := config.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}
	logg.InitLoggers()
	// initialize database
	storage, err := internal.New(configuration)
	if err != nil {
		logg.ErrorLog.Fatalf("%v failed to init storage", err)
	}
	defer func(storage *sql.DB) {
		err := storage.Close()
		if err != nil {
			logg.ErrorLog.Println(err)
		}
	}(storage)
	limiter := rl.NewRateLimiter(configuration.RateLimit)
	// initialize router
	router := http.NewServeMux()
	// TLS config
	cer, err := tls.LoadX509KeyPair(configuration.CertFile, configuration.KeyFile)
	if err != nil {
		log.Println(err)
		return
	}
	// initialize router
	srv := &http.Server{
		Addr:           configuration.Address,
		ErrorLog:       logg.ErrorLog,
		Handler:        router,
		IdleTimeout:    time.Second * configuration.IdleTimeout,
		WriteTimeout:   time.Second * configuration.WriteTimeout,
		ReadTimeout:    time.Second * configuration.ReadTimeout,
		MaxHeaderBytes: configuration.MaxHeaderBytes,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{cer},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			},
			MinVersion: tls.VersionTLS12,
		},
	}
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/main", http.StatusMovedPermanently)
	})
	router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		handlers.RegisterHandler(w, r, storage, limiter)
	})
	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		handlers.LoginHandler(w, r, storage, limiter)
	})
	router.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		handlers.LogoutHandler(w, r, storage, limiter)
	})
	router.HandleFunc("/main", func(w http.ResponseWriter, r *http.Request) {
		handlers.MainPageHandler(w, r, storage, limiter)
	})
	router.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
		handlers.SinglePostHandler(w, r, storage, limiter)
	})
	router.HandleFunc("/add-post", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddPostHandler(w, r, storage, limiter)
	})
	router.HandleFunc("/add-comment", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddCommentHandler(w, r, storage, limiter)
	})
	router.HandleFunc("/add-reaction-to-post", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddReactionToPostHandler(w, r, storage, limiter)
	})
	router.HandleFunc("/add-reaction-to-comment", func(w http.ResponseWriter, r *http.Request) {
		handlers.AddReactionToCommentHandler(w, r, storage, limiter)
	})
	router.HandleFunc("/login-with-google", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleGoogleLogin(w, r, storage, limiter)
	})
	router.HandleFunc("/oauth2callback-google", handlers.HandleGoogleCallback(handlers.GoogleClientID, handlers.GoogleClientSecret, handlers.GoogleTokenURL, storage))
	router.HandleFunc("/login-with-github", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandleGithubLogin(w, r, storage, limiter)
	})
	router.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		handlers.Admin(w, r, storage, limiter)
	})
	router.HandleFunc("/moderation-request", func(w http.ResponseWriter, r *http.Request) {
		handlers.CheckModerRequests(w, r, storage, limiter)
	})
	router.HandleFunc("/update-user", func(w http.ResponseWriter, r *http.Request) {
		handlers.UpdateUser(w, r, storage, limiter)
	})
	router.HandleFunc("/delete-post", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeletePost(w, r, storage, limiter)
	})
	router.HandleFunc("/report-post", func(w http.ResponseWriter, r *http.Request) {
		handlers.ReportPost(w, r, storage, limiter)
	})
	router.HandleFunc("/delete-tag", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteTag(w, r, storage, limiter)
	})
	router.HandleFunc("/create-tag", func(w http.ResponseWriter, r *http.Request) {
		handlers.CreateTag(w, r, storage, limiter)
	})
	router.HandleFunc("/report-status", func(w http.ResponseWriter, r *http.Request) {
		handlers.ReportStatus(w, r, storage, limiter)
	})
	router.HandleFunc("/action", func(w http.ResponseWriter, r *http.Request) {
		handlers.Actions(w, r, storage, limiter)
	})
	router.HandleFunc("/edit-post", func(w http.ResponseWriter, r *http.Request) {
		handlers.EditPost(w, r, storage, limiter)
	})
	router.HandleFunc("/edit-comment", func(w http.ResponseWriter, r *http.Request) {
		handlers.EditComment(w, r, storage, limiter)
	})
	router.HandleFunc("/delete-comment", func(w http.ResponseWriter, r *http.Request) {
		handlers.DeleteComment(w, r, storage, limiter)
	})
	router.HandleFunc("/oauth2callback-github", handlers.HandleGithubCallback(handlers.GithubClientID, handlers.GithubClientSecret, handlers.GithubTokenURL, storage))
	router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	router.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	log.Printf("Server is running on https://%s", configuration.Address)
	err = srv.ListenAndServeTLS("", "")
	if err != nil {
		fmt.Println("hah")
		logg.ErrorLog.Println(err)
		return
	}
}
