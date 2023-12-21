package tmpl

import (
	logg "forum-project/internal/forum_logger"
	"html/template"
	"net/http"
)

func RenderTemplate(w http.ResponseWriter, page string, data interface{}) {
	tmpl, err := template.ParseGlob("web/templates/*")
	if err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, page, data)
	if err != nil {
		logg.ErrorLog.Println(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
