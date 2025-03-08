package main

import (
	"awesomeProject/config"
	maincontroller "awesomeProject/controllers/main"
	parsercontroller "awesomeProject/controllers/parser"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func main() {
	r := httprouter.New()
	routes(r)

	var err error

	/*config.DB, err = common.GetMysql()

	if err != nil {
		log.Fatal("Can't open mysql:", err)

		return
	}

	common.InitRepo()
	common.UpdateRepo(config.DB)*/

	defer config.DB.Close()

	err = http.ListenAndServe("localhost:80", r)
	if err != nil {
		log.Fatal(err)
	}
}

func routes(r *httprouter.Router) {
	r.GET("/", maincontroller.StartPage)
	r.GET("/auth", maincontroller.GetToken)
	r.GET("/video", parsercontroller.GetVideo)
	r.POST("/parser", parsercontroller.Parser)
	r.POST("/getLink", parsercontroller.GetCDN)
}
