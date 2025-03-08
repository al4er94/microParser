package common

import (
	"awesomeProject/config"
	"awesomeProject/repo"
	"fmt"
	"github.com/SevereCloud/vksdk/v2/object"
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

	if _, ok := repo.RepoVkId[video.ID]; ok {
		log.Println("skip video Id download: ", video.ID, " just update")
		UpdateEntity(video)

		return
	}

	player.Host = "m.vk.com"
	params, err := url.ParseQuery(player.RawQuery)

	if err != nil {
		log.Println(err)

		return
	}

	finishUrl := player.Scheme + "://" + player.Host + "/video" + params.Get("oid") + "_" + params.Get("id")

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

	saveContentIntoDb(video, videoId, finishUrl, imgPath)
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

func DownloadPhoto(filepath, videoId, extension, url string) (string, error) {

	filepath = filepath + pathDelimetr + videoId

	ex, err := exists(filepath)

	if err != nil || !ex {
		log.Fatal("err check file path: ", err)
	}

	filepath = filepath + pathDelimetr + videoId + "_" + extension + ".jpg"

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
