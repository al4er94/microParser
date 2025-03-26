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

func GetTestContent(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	videoUrlEncode := r.FormValue("v")

	urlVideo, err := base64.StdEncoding.DecodeString(videoUrlEncode)

	if err != nil {
		log.Println("error base64 decode:", err)

		return
	}

	client := &fasthttp.Client{}
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(string(urlVideo))
	req.Header.SetUserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 YaBrowser/25.2.0.0 Safari/537.36")
	//req.Header.Set("Accept-Language", "ru-RU")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Cookie", "remixstlid=9081961991325732841_HbRjU6cK8wgu5LjqKqZKKMOcwZNRZbR9Q4TkmdfqTxP; remixstid=861545805_GciYmGTGbygspxvIbRe3ej109iX8WdygDdYDPsI2ITw; remixdark_color_scheme=0; remixcolor_scheme_mode=auto; remixdt=0; tmr_lvid=7930ec8eef42cfd9be7e7d805e5fc291; tmr_lvidTS=1714581468112; remixuas=NjFiYWEyNTVlMWM1NjU2YWE2OTQwZTA3; remixscreen_dpr=1; remixscreen_depth=24; remixscreen_winzoom=1; remixvideo_mvk_app_promo_slow_yt_modal=%7B%22lastSeenTimestamp%22%3A1728839745%2C%22showCount%22%3A1%7D; remixscreen_width=1920; remixscreen_height=1080; remixuacck=ff5df18f1d34ee29e3; remixsuc=1%3A; remixmdevice=1920/1080/1/!!-!!!!!!!!/1920; remixpuad=4BaoH2G1Qg32L_YimcMMZEzNLtfgINZe_jbk-RVt2MY; remixnsid=vk1.a.xmuMQYkak_LE4fF7FA6k3Ab5rkavM1SRHdJFNUA9uwPTfgSbOstObHggPTE7R8mwqr26bEUfFe3fn12br0gtc39w-iGyOLqmGW7sc9ZNUJ35B7N8NlIvWgeYX1v5MQKZHZRKulLtVPQMEJ_fnMOubOUFq5nvOWRihPT9dFDHweCHY1myRF0cWXEvDeswsPri; remixrefkey=2ceb63702a45001577; remixlgck=f11eaa84330dbd7cc4; remixua=43%7C-1%7C180%7C4028048701; remixff=10111111111111; hitw429=1; remixmvk-fp=24eca42e1d32a43b2811316345cfee1a; remixgp=fa7078a3691f0b768e7af1fda20563e2; remixcurr_audio=undefined; domain_sid=FnZgMBAkmxVRPQRdisqPu%3A1743010915361; tmr_detect=0%7C1743010952889; remixmsts=%7B%22data%22%3A%5B%5B1743010953%2C%22unique_adblock_users%22%2C0%2C%22mvk%22%2C%22false%22%2Cnull%2C%22video%22%5D%5D%2C%22uniqueId%22%3A489627805.70618325%7D")
	req.Header.SetMethod(fasthttp.MethodGet)

	resp := fasthttp.AcquireResponse()

	err = client.Do(req, resp)

	if err != nil {
		log.Println("err file: ", err)

		return
	}

	code := string(resp.Body())

	fmt.Println(code)

	urlsMap := common.GetUrls(code)

	var data VideoData

	urlCode := ""

	for extention, urlVal := range urlsMap {
		if extention == common.Extention360 {
			urlCode = urlVal
		}
	}

	str := base64.StdEncoding.EncodeToString([]byte(urlCode))

	fmt.Println(urlCode)

	data.Url = "/video?v=" + str + "&t=123sdqqwe"
	data.ImgUrl = ""
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
