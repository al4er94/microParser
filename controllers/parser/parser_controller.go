package parser

import (
	"awesomeProject/common"
	"awesomeProject/config"
	"awesomeProject/repo"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/api"
	"github.com/julienschmidt/httprouter"
	"github.com/valyala/fasthttp"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

const ttlLink = 600

const host = "http://site/"

type VideoData struct {
	Url    string
	ImgUrl string
	Width  string
	Height string
}

type token struct {
	Time int64  `json:"time"`
	Ua   string `json:"ua"`
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
	count := 40
	offset := 0
	go func() {
		for i := 0; i <= 10; i++ {
			time.Sleep(2 * time.Second)

			vidParams := api.Params{}
			vidParams["owner_id"] = "-" + strconv.Itoa(group[0].ID)
			vidParams["offset"] = offset
			vidParams["count"] = count
			// get information about the group
			videos, err := vk.VideoGet(vidParams)

			if err != nil {
				log.Fatal(err)
			}

			for _, video := range videos.Items {
				time.Sleep(1 * time.Second)
				common.ParseVideo(video)
			}

			offset += count
		}
	}()

	fmt.Fprint(rw, "save files..")
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

func GetContent(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	tkn := r.FormValue("tkn")

	var tknStruct token

	jsonTkn, err := base64.StdEncoding.DecodeString(tkn)

	if err != nil {
		log.Println("error base64 decode:", err)

		return
	}

	log.Println(string(jsonTkn))

	err = json.Unmarshal(jsonTkn, &tknStruct)

	if err != nil {
		log.Println("error json decode:", err)

		return
	}

	videoId := r.FormValue("id")

	id, err := strconv.Atoi(videoId)
	if err != nil {
		fmt.Println("err strconv: ", err)

		return
	}

	if !validateToken(r, tknStruct) {
		fmt.Println("err validate token: ", err)

		return
	}

	entity, ok := repo.RepoUrl[id]

	if !ok {
		log.Println("err video id: ", id)
	}

	client := &fasthttp.Client{}
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(entity.Url)
	req.Header.SetUserAgent(r.UserAgent())
	//req.Header.Set("Accept-Language", "ru-RU")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Cookie", "remixmdevice=2560/1440/2/!!-!!!!!!!!-/2560;")
	req.Header.SetMethod(fasthttp.MethodGet)

	resp := fasthttp.AcquireResponse()

	err = client.Do(req, resp)

	if err != nil {
		log.Println("err file: ", err)

		return
	}

	code := string(resp.Body())

	urlsMap := common.GetUrls(code)

	var data VideoData

	urlCode := ""

	for extention, urlVal := range urlsMap {
		if extention == common.Extention360 {
			urlCode = urlVal
		}
	}

	str := base64.StdEncoding.EncodeToString([]byte(urlCode))

	data.Url = "/video?v=" + str + "&t=123sdqqwe"
	data.ImgUrl = host + entity.ImgUrl
	data.Width = r.FormValue("width")
	data.Height = r.FormValue("height")

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

func validateToken(r *http.Request, tkn token) bool {
	if r.UserAgent() != tkn.Ua {
		log.Println("diff UA")

		return false
	}

	now := time.Now().Unix()

	if now-tkn.Time > ttlLink {
		log.Println("ttl exp")

		return false
	}

	return true
}
