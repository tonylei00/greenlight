package main

import (
	"fmt"
	"net/http"
)

// GET /v1/healthcheck
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "status: avaliable")
	fmt.Fprintf(w, "version: %v\n", version)
	fmt.Fprintf(w, "env: %v\n", app.config.env)
}
