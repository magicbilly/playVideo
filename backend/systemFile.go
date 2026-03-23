package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

func checkPosterFile(posterName, posterPath string) bool {
	posterPath = filepath.Join(posterPath, posterName)
	_, err := os.Stat(posterPath)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	} else {
		log.Err(err).Msg("poster check fail")
	}
	return false
}

type processsNum struct {
}

func insertInitData(c *Config, db *sql.DB) {
	dir, err := os.ReadDir(c.Server.Path)
	a := filepath.Join(c.Server.Path, "poster") //这个是poster路径
	aerr := os.MkdirAll(a, 0755)
	limitchan := make(chan processsNum, 3)
	if aerr != nil {
		log.Fatal().Err(aerr).Msg("")
	}
	if err != nil {
		return
	}
	for _, name := range dir {
		log.Info().Msg("开始插入数据并解析")
		title := name.Name()
		filehash := getStorageName(title)
		ext := filepath.Ext(title)
		name1 := title                         //视频路径
		title = strings.TrimSuffix(title, ext) //视频名称
		posterPath := title + ".png"           //封面路径
		if ext == ".mkv" && checkData(db, filehash) == false {
			err = insertData(db, title, name1, posterPath, filehash, 0)
		} else {
			continue
		}
	}
	noSplitVideo := checkStatus(db)
	for _, i := range noSplitVideo {
		go func(id int) {
			limitchan <- processsNum{}
			i2(db, c, id)
			defer func() { <-limitchan }()
		}(i)
	}
}
func checkStatus(db *sql.DB) []int {
	row := `select id from video where status=0`
	rows, err := db.Query(row)
	if err != nil {

	}
	defer rows.Close()
	var a []int
	for rows.Next() {
		var w int
		err := rows.Scan(&w)
		if err != nil {

		}
		a = append(a, w)
	}
	return a
}
func getvideo(db *sql.DB, a int) Video {
	row := `select FileHash,Title,Path,Poster from video where id=?`
	var video Video
	db.QueryRow(row, a).Scan(&video.Filehash, &video.Title, &video.Path, &video.Poster)
	log.Info().Msg("获取未处理视频成功")
	return video
}
func i2(db *sql.DB, c *Config, id int) {
	video := getvideo(db, id)
	fmt.Println(video)
	var vp VideoProcessor
	vp.init(c.Server.Path, video.Poster, video.Title, video.Filehash)
	fmt.Println(vp)
	log.Info().Msg("开始处理")
	err := vp.mkdirVideo(db, id)
	if err != nil {
		log.Error().Err(err).Msg("mkdirVideo 失败")
		return
	}
	if !checkPosterFile(video.Poster, filepath.Join(c.Server.Path, "poster")) {
		err := vp.generatePoster() //制作封面
		if err != nil {
			log.Err(err).Msg("封面制作失败")
		}
	}
	log.Info().Msg(video.Title + "解析完成")
}
