package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("the requested record does not exist")
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
