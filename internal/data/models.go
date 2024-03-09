package data

import (
	"database/sql"
	"errors"
)

var ErrRecordNotFound = errors.New("record not found")
var ErrEditConflict = errors.New("edit conflict")

type Models struct {
	Users       UserModel
	Players     PlayerModel
	Tokens      TokenModel
	Permissions PermissionModel
}

func NewModels(initDb *sql.DB) Models {
	return Models{
		Users:       UserModel{db: initDb},
		Players:     PlayerModel{db: initDb},
		Tokens:      TokenModel{db: initDb},
		Permissions: PermissionModel{db: initDb},
	}
}
