package common

import (
	"awesomeProject/repo"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
)

type video struct {
	id         string `field:"id"`
	vkId       string `field:"vkId"`
	url        string `field:"url"`
	previewUrl string `field:"previewUrl"`
}

func GetMysql() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/parser")

	if err != nil {
		return nil, err
	}

	return db, nil
}

func UpdateRepo(db *sql.DB) {
	rows, err := db.Query("SELECT id, vkId, url, previewUrl FROM video_contents")

	if err != nil {
		fmt.Println("can't select mysql: ", err)

		return
	}

	for rows.Next() {
		v := video{}

		err := rows.Scan(&v.id, &v.vkId, &v.url, &v.previewUrl)

		if err != nil {
			fmt.Println("Err scan: ", err)

			continue
		}

		vkId, err := strconv.Atoi(v.vkId)
		if err != nil {
			fmt.Println("err strconv: ", err)

			continue
		}

		id, err := strconv.Atoi(v.id)
		if err != nil {
			fmt.Println("err strconv: ", err)

			continue
		}

		var entity repo.VideoEntity

		entity.Url = v.url
		entity.ImgUrl = v.previewUrl

		repo.RepoVkId[vkId] = struct{}{}
		repo.RepoUrl[id] = entity
	}
}
