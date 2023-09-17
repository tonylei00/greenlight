package main

import (
	"fmt"
	"net/http"
)

// POST /v1/movie
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Creating a movie...")
}

// GET /v1/movie/:id
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	fmt.Fprintf(w, "showing info for movie id: %d\n", id)
}
