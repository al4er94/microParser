package maincontroller

import (
	"awesomeProject/config"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"net/http"
	"path/filepath"
)

type IndexData struct {
	Token    string
	AuthLink string
	ShowLink int
}

func StartPage(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	var data IndexData
	data.Token = config.Token
	data.ShowLink = 0
	data.AuthLink = "https://oauth.vk.com/authorize?client_id=51908038&display=page&redirect_uri=http://localhost:81/auth&scope=video,groups&response_type=token&v=5.131&state=123456"

	if len(data.Token) == 0 {
		data.Token = "empty token"
		data.ShowLink = 1
	}

	path := filepath.Join("public", "html", "main", "index.html")
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		http.Error(rw, err.Error(), 400)

		return
	}

	err = tmpl.ExecuteTemplate(rw, "index", data)
	if err != nil {
		http.Error(rw, err.Error(), 400)

		return
	}
}

func GetToken(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	token := r.FormValue("access_token")

	if len(token) != 0 {
		config.Token = token
		http.Redirect(rw, r, "/", http.StatusFound)
	}

	path := filepath.Join("public", "html", "main", "token.html")
	tmpl, err := template.ParseFiles(path)
	if err != nil {
		http.Error(rw, err.Error(), 400)

		return
	}

	rw.Header()
	err = tmpl.Execute(rw, nil)
	if err != nil {
		http.Error(rw, err.Error(), 400)

		return
	}
}
