package common

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"strconv"
)

type video struct {
	vkId string
}

var Repo map[int]struct{}

func GetMysql() (*sql.DB, error) {
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/site")

	if err != nil {
		return nil, err
	}

	return db, nil
}

func InitRepo() {
	Repo = make(map[int]struct{})
}

func UpdateRepo(db *sql.DB) {
	rows, err := db.Query("SELECT vkId FROM video_contents")

	if err != nil {
		log.Fatal("can't select mysql: ", err)

		return
	}

	for rows.Next() {
		v := video{}

		err := rows.Scan(&v.vkId)

		if err != nil {
			fmt.Println("Err scan: ", err)

			continue
		}

		vkId, err := strconv.Atoi(v.vkId)
		if err != nil {
			fmt.Println("err strconv: ", err)

			continue
		}

		Repo[vkId] = struct{}{}
	}
}
