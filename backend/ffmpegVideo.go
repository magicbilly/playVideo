package main

import (
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

type ff interface {
	init(BaseDir, PosterDir, title, filehash string)
	mkdirVideo(db *sql.DB, id int, tspath string) error
	generatePoster() error
}

// VideoProcessor 处理视频转换逻辑
type VideoProcessor struct {
	BaseDir   string // 配置文件中的基础视频路径
	PosterDir string // 封面存储路径
	Title     string //title是指定文件名称，没有任何后缀
	Filehash  string //文件名的哈希值
}

func (vp *VideoProcessor) init(BaseDir, PosterDir, title, filehash string) {
	vp.BaseDir = BaseDir
	vp.PosterDir = PosterDir
	vp.Title = title
	vp.Filehash = filehash
}

// videoPath是纯达到配置文件中指定的路径
// title是指定文件名称，没有任何后缀
func (vp *VideoProcessor) mkdirVideo(db *sql.DB, id int, tspath string) error {
	row := `update video set status=1 where id=?`
	_, err := db.Exec(row, id)
	var videoM3u8 string
	if tspath == "default" {
		videoM3u8 = filepath.Join(vp.BaseDir, "m3u8")
		videoM3u8 = filepath.Join(videoM3u8, vp.Filehash)
	} else {
		//直接插在下面的输出
		videoM3u8 = filepath.Join(tspath, vp.Filehash)
	}
	f := vp.Title + ".mkv"
	videoTitle := filepath.Join(vp.BaseDir, f)
	err = os.MkdirAll(videoM3u8, 0755)
	if err != nil {
		log.Error().Msg("创建文件失败")
		return err
	}
	cmd := exec.Command("ffmpeg", "-i", videoTitle, "-c:v", "libx264", "-c:a", "aac", "-f", "hls", "-hls_time", "6", "-hls_list_size", "0", filepath.Join(videoM3u8, "index.m3u8"))
	err = cmd.Run()
	if err != nil {
		row = `update video set status=3 where id=?`
		_, err = db.Exec(row, id)
	}
	row = `update video set status=2 where id=?`
	_, err = db.Exec(row, id)
	log.Info().Msg(vp.Title + "切片完成")
	return nil
}
func (vp *VideoProcessor) generatePoster() error {
	inputPath := filepath.Join(vp.BaseDir, vp.Title+".mkv")

	// 封面文件命名建议使用 hash.jpg，避免原始文件名中的特殊字符
	posterName := vp.Filehash + ".png"
	outputPath := filepath.Join(vp.PosterDir, posterName)
	log.Info().Str("input", inputPath).Msg("生成视频封面")
	// -ss 放在 -i 前面会快很多（快速定位）
	cmd := exec.Command("ffmpeg",
		"-ss", "00:00:05",
		"-i", inputPath,
		"-frames:v", "1",
		"-q:v", "2", // 图片质量 (2-5 较好)
		"-y", outputPath,
	)
	if err := cmd.Run(); err != nil {
		log.Error().Err(err).Msg("生成封面失败," + err.Error())
	}
	log.Info().Msg(vp.Title + "封面生成完成")
	return nil
}
