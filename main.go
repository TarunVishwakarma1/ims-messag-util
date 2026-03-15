package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/notification/{notificationType}", notification)
	http.ListenAndServe(":8080", r)
}

func notification(w http.ResponseWriter, r *http.Request) {

}
