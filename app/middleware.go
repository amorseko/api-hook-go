package app

import (
	"github.com/rs/cors"
	"net/http"
)

var CorsMiddleware = func(next http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"OPTIONS", "GET", "POST"},
		AllowedHeaders: []string{"*"},
		Debug:          false,
	}).Handler(next)
}
