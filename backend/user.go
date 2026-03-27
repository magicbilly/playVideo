package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"errors"

	"github.com/rs/zerolog/log"
)

type User struct {
	name     string
	id       string
	role     string
	password string
}

// UserManager 全局用户管理
type UserManager interface {
	UserInit(name, role, password string)
	UserExists(db *sql.DB, name, id string) (bool, error)
	UserUpdate(db *sql.DB, name, role, password string) error
	UserAdd(db *sql.DB) error
	UserDel(db *sql.DB) error
}

func (user *User) UserExists(db *sql.DB, name, id string) (bool, error) {
	var row string
	var ex bool
	if id != "" {
		row = `select exists (select 1 from user where id=?)`
		err := db.QueryRow(row, id).Scan(&ex)
		if err != nil {
			log.Error().Err(err).Msg("查找出错，请检查数据库")
		}
	} else {
		row = `select exists (select 1 from user where name=?)`
		err := db.QueryRow(row, name).Scan(&ex)
		if err != nil {
			log.Error().Err(err).Msg("查找出错，请检查数据库")
		}
	}
	return ex, nil
}
func (user *User) UserUpdate(db *sql.DB, name, role, password string) error {
	var row string
	if name != "" {
		user.name = name
		row = `update user set name=? where id=? `
		_, _ = db.Exec(row, name, user.id)
	}
	if role != "" {
		user.role = role
		row = `update user set role=? where id=? `
		_, _ = db.Exec(row, role, user.id)
	}
	if password != "" {

		hash := md5.New()
		hash.Write([]byte(password))
		password = hex.EncodeToString(hash.Sum(nil))
		row = `update user set password=? where id=? `
		_, _ = db.Exec(row, password, user.id)
	}
	return nil
}
func (user *User) UserAdd(db *sql.DB) error {
	exists, err := user.UserExists(db, "", user.id)
	if err != nil {
		return err
	}
	if exists == true {
		return errors.New("UserExists")
	}
	hash := md5.New()
	hash.Write([]byte(user.password))
	password := hex.EncodeToString(hash.Sum(nil))
	row := `insert into user(id,name,role,password) value(?,?,?,?)`
	_, err = db.Exec(row, user.id, user.name, user.role, password)
	if err != nil {
		return err
	}
	return nil
}
func (user *User) UserDel(db *sql.DB) error {
	row := `delete from user where id=?`
	_, err := db.Exec(row, user.id)
	if err != nil {
		return err
	}
	return nil
}
func (user *User) UserInit(name, role, password string) {
	user.name = name
	user.role = role
	user.password = password
	hash := md5.New()
	hash.Write([]byte(name + password))
	user.id = hex.EncodeToString(hash.Sum(nil))[:10]
}
