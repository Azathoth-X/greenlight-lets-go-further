package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("the requested record does not exist")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Movies MovieModel
	Users  UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{
			DB: db,
		},
		Users: UserModel{
			DB: db,
		},
	}
}
