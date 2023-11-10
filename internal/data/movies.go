package data

import (
	"database/sql"
	"github.com/lib/pq"
	"greenlight.twd.net/internal/validator"
	"time"
)

// MovieModel defines struct type which wraps a sql.DB connection pool
type MovieModel struct {
	DB *sql.DB
}

// Insert inserting a new record into the movies table
func (m MovieModel) Insert(movie *Movie) error {
	// define sql query for inserting new movie records
	query := `
		INSERT INTO movies (title, year, runtime, genres)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, version`

	// create an args slice containing the values for the placeholder params
	// from the movie struct. Declaring this slice immediately next to our SQL query
	// helps to make it nice and clear *what values are being used where* in the query.
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// use QueryRow() method to execute the SQL query on our local connection pool
	// passing in the args slice as a variadic parameter and scanning the system
	// generated id, created_at, and version values into the movie struct
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// Get a record from the movies table by its ID
func (m MovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

// Update updates an existing record in the movies table
func (m MovieModel) Update(movie *Movie) error {
	return nil
}

// Delete remove a record from the movies table by its ID
func (m MovieModel) Delete(id int64) error {
	return nil
}

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
