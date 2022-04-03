package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

//Data Model
type User struct {
	ID uint64
	Name string
	//...
}

func init() {
	var err error
	db, err = sql.Open("mysql", "root:password@tcp(192.168.77.100:3306)/test_go?charset=utf8")
	if err != nil {
		panic(err)
	}
}

func IsNoRows(err error) bool {
	if err == sql.ErrNoRows {
		return true
	}
	return false
}

func IsOtherError(err error) bool {
	if err != nil && err != sql.ErrNoRows {
		return true
	}
	return false
}

func GetUserByName(name string) (User, error){
	user := User{}
	row := db.QueryRow("select id, name from user where name = ?", name) 
	err := row.Scan(&user.ID, &user.Name)
	return user, err
}


