package data

import (
	"database/sql"
	"errors"
)

var ErrRecordNotFound = errors.New("record not found")
var ErrEditConflict = errors.New("edit conflict")

type Models struct {
	Players PlayerModel
}

func NewModels(initDb *sql.DB) Models {
	return Models{
		Players: PlayerModel{db: initDb},
	}
}
