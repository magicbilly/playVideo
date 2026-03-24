package main

import (
	"crypto/md5"
	"database/sql"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

func Initdb(config *Config) *sql.DB {
	var cfg = mysql.Config{
		User:                 config.Database.User,
		Passwd:               config.Database.Passwd,
		Net:                  config.Database.Net,
		Addr:                 config.Database.Addr,
		DBName:               config.Database.DBName,
		ParseTime:            config.Database.ParseTime,
		AllowNativePasswords: config.Database.AllowNativePasswords,
	}
	dns := cfg.FormatDSN()
	return Conndata(dns)
}
func Conndata(dns string) *sql.DB {
	db, err := sql.Open("mysql", dns)
	if err != nil {
		log.Fatal().Err(err).Msg("the database open fail,check dns")
	} else {
		log.Info().Msg("database connect success")
	}
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(time.Hour)
	perr := db.Ping()
	if perr != nil {
		log.Fatal().Err(perr).Msg("database ping fail,check network with database")
	}
	return db
}

func checkData(db *sql.DB, hash string) bool {
	query := `select exists(select 1 from video where FileHash=?)`
	var exist bool
	err := db.QueryRow(query, hash).Scan(&exist)
	if err != nil {
		log.Error().Err(err).Msg("数据库查询异常")
		return false
	}
	return exist
}

// getStorageName 将原始长文件名转换为 16 位 MD5 哈希值
func getStorageName(originalName string) string {
	// 1. 去掉首尾空格，防止因为空格导致同一个文件生成不同 Hash
	cleanName := strings.TrimSpace(originalName)

	// 2. 创建 MD5 哈希器
	hasher := md5.New()
	hasher.Write([]byte(cleanName))

	fullHash := string(hasher.Sum(nil))
	return fullHash
}

func insertData(db *sql.DB, title, path, poster, filehash string, status int) error {
	insql := `insert into video(Title,Path,Poster,FileHash,status) values(?,?,?,?)`
	_, err := db.Exec(insql, title, path, poster, filehash)
	return err
}

func GetVideo(db *sql.DB) (videos []Video, err error) {
	row := `select Title,Path,Poster,FileHash from video`
	rows, err := db.Query(row)
	if err != nil {
		log.Err(err).Msg("videos get fail")
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Err(err).Msg("database query connection close fail")
		}
	}(rows)
	for rows.Next() {
		var video Video
		err := rows.Scan(&video.Title, &video.Path, &video.Poster, &video.Filehash)
		if err != nil {
			return nil, err
		}
		videos = append(videos, video)
	}
	return videos, nil
}
func fuzzySearch(db *sql.DB, rows, title string) (videos []Video) {
	query, err := db.Query(rows, title)
	if err != nil {
		return
	}
	defer query.Close()
	var video Video
	for query.Next() {
		query.Scan(&video.Title, &video.Path, &video.Poster)
		videos = append(videos, video)
	}
	return videos
}
