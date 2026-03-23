package main

import (
	"database/sql"
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
		go func(name os.DirEntry) {
			limitchan <- processsNum{}
			i(db, name, c)
			defer func() { <-limitchan }()

		}(name)
	}
}

func i(db *sql.DB, name os.DirEntry, c *Config) {
	title := name.Name()
	filehash := getStorageName(title)
	ext := filepath.Ext(title)
	if ext == ".mkv" && checkData(db, filehash) == false {
		log.Info().Msg("开始插入数据并解析")
		name1 := title                         //视频路径
		title = strings.TrimSuffix(title, ext) //视频名称
		posterPath := title + ".png"           //封面路径
		var vp ff
		vp.init(c.Server.Path, posterPath, title, filehash)
		err := vp.mkdirVideo()
		if err != nil {
			return
		}
		if !checkPosterFile(posterPath, filepath.Join(c.Server.Path, "poster")) {
			err := vp.generatePoster() //制作封面
			if err != nil {
				log.Fatal().Err(err).Msg("封面制作失败")
			}
		}
		err = insertData(db, title, name1, posterPath, filehash)
		log.Info().Msg(title + "解析完成")
		if err != nil {
			return
		}
	}
}

//以下函数已经移动到了ffmpegVideo.go中ff接口作为方法使用
// func generatePoster(videoPath string, posterPath string) error { //封面制作
//
//		// 执行指令：ffmpeg -i 视频路径 -ss 1 -frames:v 1 封面路径
//
//		cmd := exec.Command("ffmpeg", "-i", videoPath, "-ss", "00:00:10", "-frames:v", "1", "-y", posterPath)
//
//		// -y 表示如果封面已存在则自动覆盖
//		err := cmd.Run()
//		if err != nil {
//			log.Err(err).Msg("change poster fail")
//		}
//		return err
//	}
// videoPath是纯达到配置文件中指定的路径
// title是指定文件名称，没有任何后缀
//func mkdirVideo(videoPath, Filehash, Title string) error {
//	videoM3u8 := filepath.Join(videoPath, "m3u8")
//	videoM3u8 = filepath.Join(videoM3u8, Filehash)
//	Title = Title + ".mkv"
//	videoTitle := filepath.Join(videoPath, Title)
//	err := os.MkdirAll(videoM3u8, 0755)
//	if err != nil {
//		log.Error().Msg("创建文件失败")
//		return err
//	}
//	cmd := exec.Command("ffmpeg", "-i", videoTitle, "-c:v", "libx264", "-c:a", "aac", "-f", "hls", "-hls_time", "6", "-hls_list_size", "0", filepath.Join(videoM3u8, "index.m3u8"))
//	//	cmd.Stdout = os.Stdout
//	//	cmd.Stderr = os.Stderr
//
//	err = cmd.Run()
//	if err != nil {
//		//		fmt.Println(cmd.Stdout)
//		//		fmt.Println(cmd.Stderr)
//		return err
//	}
//	return nil
//}
