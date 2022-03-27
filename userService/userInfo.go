package userService

import (
	"LessionGo/db"
	"github.com/pkg/errors"
)

type UserInfo struct {
	ID uint64
	Name string
	//...
}

func GetUserInfo(name string) (UserInfo, error){
	var userInfo UserInfo
	user, err := db.GetUserByName(name)
	switch {
	case err == nil:
		userInfo.ID = user.ID
		userInfo.Name = user.Name
	case db.IsNoRows(err):
		err = errors.Wrapf(err, "User %q Unexisted.", name)
	case db.IsOtherError(err):
		err = errors.Wrapf(err, "Other Error Happened.")
	}
	return userInfo, err
}