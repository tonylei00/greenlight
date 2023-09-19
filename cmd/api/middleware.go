package main

import (
	"fmt"
	"net/http"
)

// Close the connection on the current goroutine and send back a 500 when the stack unwinds in the event of a panic
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// "Connection": "close" header will auto trigger the connection to close
				w.Header().Set("Connection", "close")
				// Note: Need to interpolate the panic err because it has type any
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
