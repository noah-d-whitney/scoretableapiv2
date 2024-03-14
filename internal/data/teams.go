package data

import (
	"database/sql"
	"time"
)

type Team struct {
	ID        int64
	Name      string
	Size      int
	CreatedAt time.Time
	Version   int32
	IsActive  bool
}

type TeamModel struct {
	db sql.DB
}

func (m *TeamModel) Get()
