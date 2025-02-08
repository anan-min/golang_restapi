package main

import (
	// "database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/labstack/echo/v4"
	_ "github.com/proullon/ramsql/driver"
)

// var db *sql.DB

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
	yearQueryParam := c.QueryParam("year")
	var returnedMovies []Movie

	if yearQueryParam != "" {
		year, err := strconv.Atoi(yearQueryParam)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid year"})
		}

		returnedMovies, err = getMoviesByYear(year)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to query database"})
		}

		if len(returnedMovies) == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "No movies found for the given year"})
		}

		return c.JSON(http.StatusOK, returnedMovies)
	}

	return c.JSON(http.StatusOK, movies)
}


func getMoviesByYear(year int) ([]Movie, error) {
	var returnedMovies []Movie
	// If year is -1, we want to return all movies
	if year == -1 {
		return movies, nil
	} else {
		for i := 0; i < len(movies); i++ {
			if movies[i].Year == year {
				returnedMovies = append(returnedMovies, movies[i])
			}
		}
		return returnedMovies, nil
	}

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



func main() {
	
	e := echo.New()
	// // Define routes
	e.GET("/movies", getAllMoviesHandler)
	e.GET("/movies/:id", getMovieByIdHandler)
	e.POST("/movies", createMovieHandler)

	// // Start server
	port := "2565"
	fmt.Printf("Starting server on port %s...\n", port)
	fmt.Println(e.Start(":" + port))
}
