package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	CONTENT_TYPE string = "Content-Type"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/notification/{notificationType}", notification)
	r.Post("/email", email1)
	r.Get("/", events)

	http.ListenAndServe(":8080", r)
}

func notification(w http.ResponseWriter, r *http.Request) {
	value := r.PathValue("notificationType")
	w.WriteHeader(200)
	w.Write([]byte(value))
	w.Header().Set(CONTENT_TYPE, "application/json")
}

// Works
func email1(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Erorr in getting data from req: ", err)
	}
	//email.SendEmail(body)
	fmt.Println(string(body))
	value := r.PathValue("notificationType")
	//email.SendEmail(body)
	w.WriteHeader(200)
	w.Write([]byte(value))
	w.Header().Set(CONTENT_TYPE, "application/json")

}

func events(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(CONTENT_TYPE, "text/event-stream")

	tokens := []string{"this", "is", "a", "live", "east", "test", "from", "you", "tube"}

	for _, token := range tokens {
		content := fmt.Sprintf("data: %s\n", token)
		w.Write([]byte(content))
		w.(http.Flusher).Flush()
	}
}
