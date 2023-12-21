package handlers

import (
	logg "forum-project/internal/forum_logger"
	"net/http"
	"text/template"
)

func ErrorHandler(w http.ResponseWriter, code int, message string) {
	response := struct {
		ErrorCode    int
		ErrorText    string
		ErrorMessage string
	}{
		ErrorCode:    code,
		ErrorText:    http.StatusText(code),
		ErrorMessage: message,
	}

	tmpl, err := template.ParseFiles("web/templates/error_page.html")
	if err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(code)

	err = tmpl.Execute(w, response)
	if err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
