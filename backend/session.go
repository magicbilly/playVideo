package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/rs/zerolog/log"
)

//只有单元测试完成

var private = "u9K85E0hk310nWeHba4b"
var store = sessions.NewCookieStore([]byte(private))

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}
}
func isTruePassword(name, password string, db *sql.DB) bool {
	row := `select exists(select 1 from user where name=? and password=?)`
	var isCorrect bool
	if name == "1" {
		db.QueryRow(row, name, password).Scan(&isCorrect)
		return isCorrect
	} else {
		hash := md5.New()
		hash.Write([]byte(password))
		passwordMd5 := hex.EncodeToString(hash.Sum(nil))
		db.QueryRow(row, name, passwordMd5).Scan(&isCorrect)
		return isCorrect
	}
}
func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. 获取前端传来的账号密码
		name := r.FormValue("name")
		pass := r.FormValue("password")
		log.Info().Msg(name + pass)
		// 2. 调用你的 manager 去验证并获取“画像”
		// 假设你给 manager 增加了一个校验方法
		// 或者先 Exists 再初始化
		exists := isTruePassword(name, pass, db)
		if !exists {
			http.Error(w, "用户不存在", http.StatusUnauthorized)
			return
		}

		// 3. 验证通过后，构造当前登录的 User 对象
		// 注意：这里的 u 是根据本次请求动态生成的
		u := &User{name: name}
		u.UserInit(name, "visitor", pass) // 重新计算 ID 匹配数据库
		// 获取 Session（如果不存在则创建新的）
		session, _ := store.Get(r, "video-auth")

		// 在 Session 中存储关键信息
		session.Values["user_id"] = u.id
		session.Values["user_role"] = u.role
		session.Values["is_login"] = true
		// 💡 重要：必须调用 Save，否则 Cookie 不会被发送给浏览器
		if err := session.Save(r, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"code": 200, "message": "success"}`))
		log.Info().Msgf("用户 %s 登录成功并已下发 Session", u.name)
	}
}
