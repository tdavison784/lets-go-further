package data

import (
	"database/sql"
	"errors"
)

// ErrRecordNotFound Define a custom ErrRecordNotFound error. We'll return this from our Get() method
// when looking up a movie that doesn't exist
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// Models struct which wraps around MovieModel struct. We'll add other models to this
// like a UserModel and PermissionsModel
type Models struct {
	Movies      MovieModel
	Users       UserModel
	Tokens      TokenModel
	Permissions PermissionsModel
}

// NewModels is a helper func that returns a Models struct containing
// the initialized MoviesModel
func NewModels(db *sql.DB) Models {
	return Models{
		Movies:      MovieModel{DB: db},
		Users:       UserModel{DB: db},
		Tokens:      TokenModel{DB: db},
		Permissions: PermissionsModel{DB: db},
	}
}
