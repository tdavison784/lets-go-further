package data

import (
	"greenlight.twd.net/internal/validator"
	"time"
)

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year"`
	Runtime   Runtime   `json:"runtime"`
	Genres    []string  `json:"genres"`
	Version   int32     `json:"version"`
}

// ValidateMovie function is used to run our validation checks on client user input.
// we have it here to keep more of the business logic outside our handlers
func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "title must not be empty")
	v.Check(len(movie.Title) <= 500, "title", "title must be less then 500 bytes long")

	v.Check(movie.Year != 0, "year", "must provide a year")
	v.Check(movie.Year >= 1888, "year", "year must be greater then 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "year cannot be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must provide a valid runtime")
	v.Check(movie.Runtime > 0, "runtime", "runtime must be greater than 0")

	v.Check(movie.Genres != nil, "genres", "genres must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "cannot contain more then 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "cannot contain duplicate genres")
}
