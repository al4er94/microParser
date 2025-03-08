package common

import (
	"awesomeProject/config"
	"bytes"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/object"
	"github.com/valyala/fasthttp"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const pathDelimetr = "/"

const Extention360 = "360"
const extention720 = "720"

const folderName = "videos"

//https://vk.com/video/@slivdevushek_tg/all
//https://vk.com/video/180937359/all
//https://vkvideo.ru/video-226450474_456239142
//https://vkvideo.ru/video-226450474_456240153

var extentions = []string{
	Extention360,
	//extention720,
}

func ParseVideo(video object.VideoVideo) {

	log.Println(video.Player, " title: ", video.Title, " add: ", video.Added)

	player, err := url.Parse(video.Player)

	if err != nil {
		log.Fatal(err)
	}

	videoId := strconv.Itoa(video.ID)

	if _, ok := Repo[video.ID]; ok {
		log.Println("skip video Id download: ", video.ID, " just update")
		UpdateEntity(video)

		return
	}

	player.Host = "m.vk.com"
	params, err := url.ParseQuery(player.RawQuery)

	if err != nil {
		log.Fatal(err)
	}

	finishUrl := player.Scheme + "://" + player.Host + "/video" + params.Get("oid") + "_" + params.Get("id")
	log.Println(finishUrl)

	return

	client := &fasthttp.Client{}

	req := fasthttp.AcquireRequest()
	req.SetRequestURI(finishUrl)
	req.Header.SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "ru-RU")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Cookie", "emixscreen_depth=24; remixdark_color_scheme=0; remixcolor_scheme_mode=auto; remixlang=0; remixstlid=9061487165792055558_FhTxTR2K3ZMtwXvdvyzXuqm9xBXJu3EOZNnOJjGICFP; remixstid=1543406390_aE8MNzIltcZMo4nxeiL6T9ZFPrffZd2tzwNwb6nvX0c; remixlgck=3bf57e4bc6e50bc94a; remixscreen_width=2560; remixscreen_height=1440; remixscreen_dpr=2; remixrt=1; remixdt=0; tmr_lvid=27f7e8a7eb6287faaabe85adee85d7c3; tmr_lvidTS=1708612836485; remixscreen_orient=1; remixsf=1; remixuas=YjViNzhiYzg4MDQyNWY2YmNlMTk0MmQ2; remixrefkey=51e05ffeea3f3b6a5a; remixmdevice=2560/1440/2/!!-!!!!!!!!-/2560; remixff=10111111111111; remixua=52%7C627%7C213%7C3832481977; remixscreen_winzoom=0; remixgp=2b046b05f5d04e05d038bf3a3d109cea; remixmvk-fp=176f4b3d5258a03d03a2ea7f184d22e6; remixcurr_audio=undefined; earlyhints=1; domain_sid=W4zW60v0BKjXNR-gJ-9yA%3A1714659455599; tmr_detect=0%7C1714659457714")
	req.Header.SetMethod(fasthttp.MethodGet)

	resp := fasthttp.AcquireResponse()

	err = client.Do(req, resp)

	if err != nil {
		log.Println("err: ", err)
		log.Println("can't save video: ", video.ID)

		return
	}

	code := string(resp.Body())

	urlsMap := GetUrls(code)

	videoPath := ""

	for extention, urlVal := range urlsMap {
		log.Println(extention, urlVal)

		videoPath, err = DownloadFile(folderName, videoId, extention, urlVal)

		if err != nil {
			log.Fatal("err save file: ", err)
		}

	}

	imgPath := ""

	for _, videoIm := range video.Image {
		if videoIm.Width == 720 {
			log.Println("Width: ", videoIm.Width, " IM ", videoIm.BaseImage.Width, " id: ", videoId)
			imgPath, err = DownloadPhoto(folderName, videoId, "720", videoIm.BaseImage.URL)

			if err != nil {
				log.Fatal("err save file photo: ", err)
			}
		}
	}

	saveContentIntoDb(video, videoId, videoPath, imgPath)
}
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		err = os.Mkdir(path, os.ModeDir)
		if err != nil {
			return false, err
		}

		return true, nil
	}
	return false, err
}

func GetUrls(code string) map[string]string {
	urlByExtention := map[string]string{}

	for _, extention := range extentions {
		regexpPath := "\"url" + extention + "\":\"([\\s\\S]+?)\",\""

		r, _ := regexp.Compile(regexpPath)
		url360Strings := r.FindStringSubmatch(code)
		urlPath := ""
		if len(url360Strings) == 2 {
			urlPath = url360Strings[1]
		}

		decodedValue, err := url.QueryUnescape(urlPath)
		if err != nil {
			log.Fatal(err)
		}

		decodedValue = strings.ReplaceAll(decodedValue, "\\", "")
		urlByExtention[extention] = decodedValue
	}

	return urlByExtention
}

func DownloadFile(filepath, videoId, videoExtention, url string) (string, error) {
	if len(url) == 0 {
		return "", nil
	}

	log.Println("start DownloadFile, videoId: ", videoId)

	filepath = filepath + pathDelimetr + videoId

	ex, err := exists(filepath)
	if err != nil || !ex {
		log.Fatal("err check file path: ", err)
	}

	filepath = filepath + pathDelimetr + videoId + "_" + videoExtention + ".mp4"

	log.Println("filepath: ", filepath)

	client := &fasthttp.Client{}
	out, err := os.Create(filepath + ".tmp")
	if err != nil {
		return "", err
	}

	log.Println("start request")
	req := fasthttp.AcquireRequest()
	req.Header.SetUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "ru-RU")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Cookie", "emixscreen_depth=24; remixdark_color_scheme=0; remixcolor_scheme_mode=auto; remixlang=0; remixstlid=9061487165792055558_FhTxTR2K3ZMtwXvdvyzXuqm9xBXJu3EOZNnOJjGICFP; remixstid=1543406390_aE8MNzIltcZMo4nxeiL6T9ZFPrffZd2tzwNwb6nvX0c; remixlgck=3bf57e4bc6e50bc94a; remixscreen_width=2560; remixscreen_height=1440; remixscreen_dpr=2; remixrt=1; remixdt=0; tmr_lvid=27f7e8a7eb6287faaabe85adee85d7c3; tmr_lvidTS=1708612836485; remixscreen_orient=1; remixsf=1; remixuas=YjViNzhiYzg4MDQyNWY2YmNlMTk0MmQ2; remixrefkey=51e05ffeea3f3b6a5a; remixmdevice=2560/1440/2/!!-!!!!!!!!-/2560; remixff=10111111111111; remixua=52%7C627%7C213%7C3832481977; remixscreen_winzoom=0; remixgp=2b046b05f5d04e05d038bf3a3d109cea; remixmvk-fp=176f4b3d5258a03d03a2ea7f184d22e6; remixcurr_audio=undefined; earlyhints=1; domain_sid=W4zW60v0BKjXNR-gJ-9yA%3A1714659455599; tmr_detect=0%7C1714659457714")
	req.SetRequestURI(url)
	req.Header.SetMethod(fasthttp.MethodGet)

	// Get the data
	resp := fasthttp.AcquireResponse()

	err = client.Do(req, resp)

	if err != nil {
		log.Fatal("err get: ", err)
	}

	log.Println("Status code save: ", resp.StatusCode())
	//log.Info("Status code: ", string(resp.Body()))

	defer out.Close()

	buf := []byte{}
	// Write the body to file

	log.Println("Start the body to file: ", videoId, " ext: ", extentions, "...")
	if _, err = io.Copy(out, io.TeeReader(bytes.NewBuffer(resp.Body()), bytes.NewBuffer(buf))); err != nil {
		return "", err
	}
	out.Close()
	log.Println("Finish save: ", videoId, " ext: ", extentions, "...")

	if err = os.Rename(filepath+".tmp", filepath); err != nil {
		return "", err
	}

	return filepath, nil
}

func DownloadPhoto(filepath, videoId, extension, url string) (string, error) {
	filepath = filepath + pathDelimetr + videoId + pathDelimetr + videoId + "_" + extension + ".jpg"

	log.Println("URL :", url)

	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	//open a file for writing
	file, err := os.Create(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Use io.Copy to just dump the response body to the file. This supports huge files
	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Success save photo!")

	return filepath, nil
}

func saveContentIntoDb(video object.VideoVideo, videoId, videoPath, imgPath string) {
	videoPath = "assets/" + videoPath
	imgPath = "assets/" + imgPath

	result, err := config.DB.Exec("INSERT INTO video_contents (vkId, name_ru, description_ru, url, previewUrl, views, likes) values (?, ?, ?, ?, ?, ?, ?)",
		videoId, video.Title, video.Description, videoPath, imgPath, video.Views, video.Likes.Count)

	if err != nil {
		log.Println("Err insert db: ", err)

		return
	}

	log.Println("Res: ", result)
}

func UpdateEntity(video object.VideoVideo) {
	result, err := config.DB.Exec("UPDATE video_contents SET name_ru = ? , description_ru = ?, views = ?, likes = ? WHERE vkId = ?",
		video.Title, video.Description, video.Views, video.Likes.Count, video.ID)
	if err != nil {
		log.Println("Err update db: ", err)

		return
	}

	log.Println("Res update: ", result)
}
