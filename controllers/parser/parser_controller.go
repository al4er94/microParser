package parser

import (
	"awesomeProject/common"
	"awesomeProject/config"
	"encoding/base64"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/julienschmidt/httprouter"
	"github.com/valyala/fasthttp"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

type VideoData struct {
	Url string
}

func Parser(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	link := r.FormValue("linkGroup")
	vk := api.NewVK(config.Token)

	groupParams := api.Params{}
	groupParams["group_ids"] = link
	group, err := vk.GroupsGetByID(groupParams)
	if err != nil || len(group) != 1 {
		log.Fatal(err)
	}

	log.Println(group[0].ID)

	vidParams := api.Params{}
	vidParams["owner_id"] = "-" + strconv.Itoa(group[0].ID)
	vidParams["offset"] = 200
	// get information about the group
	videos, err := vk.VideoGet(vidParams)

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		c := 0
		for _, video := range videos.Items {
			common.ParseVideo(video)
			c++
		}
	}()

	fmt.Fprint(rw, "save files..")
}

func GetCDN(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	client := &fasthttp.Client{}

	link := r.FormValue("getLink")

	req := fasthttp.AcquireRequest()
	req.SetRequestURI(link)
	req.Header.SetUserAgent(r.UserAgent())
	//req.Header.Set("Accept-Language", "ru-RU")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Cookie", "remixmdevice=2560/1440/2/!!-!!!!!!!!-/2560;")
	req.Header.SetMethod(fasthttp.MethodGet)

	resp := fasthttp.AcquireResponse()

	err := client.Do(req, resp)

	if err != nil {
		log.Println("err file: ", err)

		return
	}

	code := string(resp.Body())

	urlsMap := common.GetUrls(code)

	var data VideoData

	url := ""

	for extention, urlVal := range urlsMap {
		if extention == common.Extention360 {
			url = urlVal
		}
	}

	str := base64.StdEncoding.EncodeToString([]byte(url))
	fmt.Println(str)

	data.Url = "http://localhost/video?v=" + str + "&t=123sdqqwe"

	path := filepath.Join("public", "html", "main", "video.html")

	tmpl, err := template.ParseFiles(path)
	if err != nil {
		http.Error(rw, err.Error(), 400)

		return
	}

	err = tmpl.Execute(rw, data)

	if err != nil {
		http.Error(rw, err.Error(), 400)

		return
	}
}

func GetVideo(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	link := r.FormValue("v")

	dataCode, err := base64.StdEncoding.DecodeString(link)
	if err != nil {
		log.Fatal("error:", err)
	}
	url := string(dataCode)

	http.Redirect(rw, r, url, http.StatusFound)
}
