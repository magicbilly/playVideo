package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"runtime"
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

func getPosterPath(ppath, path string) string {
	if path == "default" {
		return filepath.Join(path, "poster")
	} else {
		return ppath
	}
}

// 这个是一个任务池类型，下面是可以限制FFmpeg协程的数量，可以在下面的make中修改协程数量
type processsNum struct {
}

func insertInitData(c *Config, db *sql.DB) {
	dir, err := os.ReadDir(c.Server.Path)
	a := getPosterPath(c.Server.Poster, c.Server.Path) //这个是poster路径
	aerr := os.MkdirAll(a, 0755)
	var court int
	if c.System.Coroutine == 0 {
		court = runtime.NumCPU() / 2
	} else {
		court = c.System.Coroutine
	}
	limitChan := make(chan processsNum, court) //限制协程数量
	if aerr != nil {
		log.Fatal().Err(aerr).Msg("")
	}
	if err != nil {
		return
	}
	for _, name := range dir {
		log.Info().Msg("开始插入数据并解析")
		title := name.Name()
		fileHash := getStorageName(title)
		ext := filepath.Ext(title)
		name1 := title                         //视频路径
		title = strings.TrimSuffix(title, ext) //视频名称
		posterPath := title + ".png"           //封面路径
		if ext == ".mkv" && checkData(db, fileHash) == false {
			err = insertData(db, title, name1, posterPath, fileHash, 0)
		} else {
			continue
		}
	}
	noSplitVideo := checkStatus(db)
	for _, i := range noSplitVideo {
		go func(id int) {
			limitChan <- processsNum{}
			i2(db, c, id)
			defer func() { <-limitChan }()
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
	var vp VideoProcessor
	vp.init(c.Server.Path, video.Poster, video.Title, video.Filehash)
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
