package data

import (
	"database/sql"
	"errors"
)

var (
	ErrorNotFoundRecord = errors.New("the requested record does not exist")
)

type Models struct {
	Movies MovieModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{
			DB: db,
		},
	}
}
