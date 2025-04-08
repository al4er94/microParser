package main

import (
	"awesomeProject/common"
	"awesomeProject/config"
	maincontroller "awesomeProject/controllers/main"
	parsercontroller "awesomeProject/controllers/parser"
	"awesomeProject/repo"
	"awesomeProject/task"
	"awesomeProject/task/videos"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var initializer = []func(){}
var finalizer = []func(){}

var initChan = make(chan struct{})
var stopChan = make(chan struct{})
var signalChan = make(chan os.Signal, 1)

func main() {
	log.Println("START SERV: ")

	r := httprouter.New()
	routes(r)

	var err error

	initTask(
		videos.NewTask(),
	)

	initialize()
	log.Println("LISTEN START ")

	go func() {
		err = http.ListenAndServe(":80", r)
		if err != nil {
			log.Fatal(err)

			os.Exit(1)
		}
	}()

	<-stopChan
	finalize()
	os.Exit(0)
}

func routes(r *httprouter.Router) {
	r.GET("/", maincontroller.StartPage)
	r.GET("/auth", maincontroller.GetToken)
	r.GET("/video", parsercontroller.GetVideo)
	r.GET("/content", parsercontroller.GetContent)
	r.GET("/test", parsercontroller.GetTestContent)
	r.POST("/parser", parsercontroller.Parser)

	r.ServeFiles("/static/*filepath", http.Dir("./public/node_modules"))
}

func initTask(t ...task.Task) {
	for _, tasks := range t {
		initializer = append(initializer, tasks.Run)
		finalizer = append(finalizer, tasks.Stop)
	}
}

func initialize() {
	var err error

	signal.Notify(signalChan, os.Interrupt)
	signal.Notify(signalChan, syscall.SIGTERM)

	go func() {
		<-signalChan
		close(stopChan)
	}()

	config.DB, err = common.GetMysql()

	if err != nil {
		log.Fatal("Can't open mysql:", err)

		return
	}

	repo.InitRepo()
	common.UpdateRepo(config.DB)

	go func() {
		for _, f := range initializer {
			f()
		}

		close(initChan)
	}()

	for {
		select {
		case <-stopChan:
			finalize()
			os.Exit(1)
		case <-initChan:
			return
		}
	}

}

func finalize() {
	log.Println("STOP")

	for _, f := range finalizer {
		f()
	}

	defer config.DB.Close()

}
