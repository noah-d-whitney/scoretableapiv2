package data

import (
	"database/sql"
	"errors"
)

var helperModels HelperModels
var ErrRecordNotFound = errors.New("record not found")
var ErrEditConflict = errors.New("edit conflict")

type Models struct {
	Users       UserModel
	Players     PlayerModel
	Teams       TeamModel
	Tokens      TokenModel
	Pins        PinModel
	Permissions PermissionModel
}

type HelperModels struct {
	Pins PinModel
}

func NewModels(initDb *sql.DB) Models {
	helperModels = HelperModels{
		Pins: PinModel{db: initDb},
	}
	return Models{
		Users:       UserModel{db: initDb},
		Players:     PlayerModel{db: initDb},
		Teams:       TeamModel{db: initDb},
		Tokens:      TokenModel{db: initDb},
		Pins:        PinModel{db: initDb},
		Permissions: PermissionModel{db: initDb},
	}
}
