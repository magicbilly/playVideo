package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port   int    `yaml:"Port"`
		Path   string `yaml:"Video_path"`
		Poster string `yaml:"Poster_path"`
	} `yaml:"Server"`
	Database struct {
		User                 string `yaml:"User"`
		Passwd               string `yaml:"Passwd"`
		Net                  string `yaml:"Net"`
		Addr                 string `yaml:"Addr"`
		DBName               string `yaml:"DBName"`
		ParseTime            bool   `yaml:"ParseTime"`
		AllowNativePasswords bool   `yaml:"AllowNativePasswords"`
	} `yaml:"Database"`
	System struct {
		Coroutine int `yaml:"Coroutine"`
	}
}
type Video struct {
	Title    string `json:"title"`
	Poster   string `json:"poster"`
	Path     string `json:"path"`
	Filehash string `json:"filehash"`
}

func corsHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func main() {
	c, err := LoadConfig("config.yaml")
	if err != nil {
		return
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	db := Initdb(c)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Error().Msg("Mysql close fail")
		}
	}(db)
	insertInitData(c, db)
	http.HandleFunc("/api/play", Index(db))
	log.Info().Msg("route '/api/play' register success")
	http.HandleFunc("/api/search", Search(db))
	log.Info().Msg("route '/api/Search' register success")
	log.Info().Msg("route '/' register success")
	// 映射你的本地视频目录到 /play/ 路径
	videoDir := c.Server.Path
	http.Handle("/play/", http.StripPrefix("/play/", http.FileServer(http.Dir(videoDir))))
	posterDir := filepath.Join(videoDir, "poster")
	http.Handle("/api/poster/", http.StripPrefix("/api/poster/", http.FileServer(http.Dir(posterDir))))
	addr := ":" + fmt.Sprintf("%d", c.Server.Port)
	//
	log.Info().Msg("the port binding")
	err = http.ListenAndServe(addr, corsHandler(http.DefaultServeMux))
	if err != nil {
		log.Fatal().Err(err).Msg("the port bind fail")
	}
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatal().Err(err).Msg("read config file fail")
	}
	var c Config
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		log.Fatal().Err(err).Msg("Config load fail")
	}
	log.Info().Msg("Config load success")
	return &c, err

}

func Index(db *sql.DB) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		log.Info().Str("method", request.Method).Str("url", request.URL.String()).Msg("收到请求")
		videos, err := GetVideo(db)
		if err != nil {
			log.Err(err).Msg("get video list fail check database")
		}
		log.Info().Msg("get video list success")
		err = json.NewEncoder(writer).Encode(videos)
		log.Info().Msg("encode success")
		if err != nil {
			log.Fatal().Err(err).Msg("get json encode fail")
		}
	}
}
func Search(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		title := "%" + r.URL.Query().Get("Title") + "%"
		rows := `select Title,Path,Poster from video where Title like ?`
		videos := fuzzySearch(db, rows, title)
		err := json.NewEncoder(w).Encode(videos)
		if err != nil {
			log.Err(err).Msg("json encode fail")
		}
	}
}
