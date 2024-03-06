package models

import "database/sql"

type Models struct {
	Players PlayerModel
}

func NewModels(initDb *sql.DB) Models {
	return Models{
		Players: PlayerModel{db: initDb},
	}
}
