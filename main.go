package main

import (
	"database/sql"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/labstack/echo/v4"
	_ "github.com/proullon/ramsql/driver"
)

var db *sql.DB

type Movie struct {
	ImdbID      string  `json:"imdbID"`
	Title       string  `json:"title"`
	Year        int     `json:"year"`
	Rating      float32 `json:"rating"`
	IsSuperHero bool    `json:"isSuperHero"`
}

var movies = []Movie{
	{ImdbID: "tt0111161", Title: "The Shawshank Redemption", Year: 1994, Rating: 9.2, IsSuperHero: false},
	{ImdbID: "tt0068646", Title: "The Godfather", Year: 1972, Rating: 9.2, IsSuperHero: false},
	{ImdbID: "tt0071562", Title: "The Dark Knight", Year: 2008, Rating: 9.0, IsSuperHero: true},
	{ImdbID: "tt0110912", Title: "The Godfather: Part II", Year: 1974, Rating: 9.0, IsSuperHero: false},
	{ImdbID: "tt0060196", Title: "12 Angry Men", Year: 1957, Rating: 8.9, IsSuperHero: false},
	{ImdbID: "tt0108052", Title: "Schindler's List", Year: 1993, Rating: 8.9, IsSuperHero: false},
	{ImdbID: "tt0073486", Title: "The Lord of the Rings: The Return of the King", Year: 2003, Rating: 8.9, IsSuperHero: true},
	{ImdbID: "tt0167260", Title: "Inception", Year: 2010, Rating: 8.8, IsSuperHero: false},
	{ImdbID: "tt1375666", Title: "The Dark Knight Rises", Year: 2012, Rating: 8.8, IsSuperHero: false},
}

func getAllMoviesHandler(c echo.Context) error {
    yearStr := c.QueryParam("year")
    year := -1 // Default value: -1 means no year filter

    if yearStr != "" {
        var err error
        year, err = strconv.Atoi(yearStr)
        if err != nil {
            return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid year format. Year must be an integer."})
        }
        if year < 0 { // Check for negative year values
            return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid year value. Year must be non-negative."})
        }
    }

    returnedMovies, err := getMoviesByYear(year)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
    }

    return c.JSON(http.StatusOK, returnedMovies)
}

func getMoviesByYear(year int) ([]Movie, error) {
    query := "SELECT imdb_id, title, year, rating, is_superhero FROM movies" 

    var args []interface{} 
    if year != -1 { 
        query += " WHERE year = ?"
        args = append(args, year) 
    }

    rows, err := db.Query(query, args...) // Use args... to pass the arguments
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var returnedMovies []Movie
    for rows.Next() {
        var movie Movie
        if err := rows.Scan(&movie.ImdbID, &movie.Title, &movie.Year, &movie.Rating, &movie.IsSuperHero); err != nil {
            return nil, err
        }
        returnedMovies = append(returnedMovies, movie)
    }

    return returnedMovies, nil
}






func getMovieByIdHandler(c echo.Context) error {
	idParam := c.Param("id")
	if idParam == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing id parameter"})
	}

	matched, err := regexp.MatchString(`^tt\d{7}$`, idParam)
	if (!matched || err != nil) {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid id format"})
	}	

	var movie Movie
	for i := 0; i < len(movies); i++ {
		if movies[i].ImdbID == idParam {
			movie = movies[i]
			return c.JSON(http.StatusOK, movie)
		}
	}

	return c.JSON(http.StatusNotFound, map[string]string{"error": "Movie not found"})
}

func createMovieHandler(c echo.Context) error {
	// parse request body
	var newMovie Movie
	if err := c.Bind(&newMovie); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request format"})
	}
	// movie params validation
	if newMovie.ImdbID == "" || newMovie.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing required fields: ImdbID or Title"})
	}
	if newMovie.Year <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid year"})
	}
	if newMovie.Rating < 0 || newMovie.Rating > 10 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Rating must be between 0 and 10"})
	}

	movies = append(movies, newMovie)
	return c.JSON(http.StatusCreated, newMovie)
}

func init() {
	conn()
	createMovieTable()
	insertAllMovies(movies)
}


func conn() {
	var err error
	db, err = sql.Open("ramsql", "goimdb")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
}

func createMovieTable() { 
	query := `
		CREATE TABLE IF NOT EXISTS movies (
		imdb_id VARCHAR(10) UNIQUE PRIMARY KEY NOT NULL,
		title VARCHAR(255) NOT NULL,
		year INT NOT NULL,
		rating FLOAT NOT NULL,
		is_superhero BOOLEAN NOT NULL
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}


func insertMovie(movie Movie) {
	query := "INSERT INTO movies (imdb_id, title, year, rating, is_superhero) VALUES (?, ?, ?, ?, ?)"
	_, err := db.Exec(query, movie.ImdbID, movie.Title, movie.Year, movie.Rating, movie.IsSuperHero)
	if err != nil {
		log.Fatal(err)
	}
}


func insertAllMovies(movies []Movie) {
	for _, movie := range movies {
		insertMovie(movie)
	}
}

func main() {
	conn()
	e := echo.New()
	// // Define routes
	e.GET("/movies", getAllMoviesHandler)
	e.GET("/movies/:id", getMovieByIdHandler)
	e.POST("/movies", createMovieHandler)

	// // Start server
	port := "2565"
	log.Printf("Starting server on port %s...\n", port)
	log.Println(e.Start(":" + port))
}
