package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

//go:embed assets/*
var assets embed.FS

func main() {
	assets, _ := fs.Sub(assets, "assets")

	indexRaw, err := fs.ReadFile(assets, "index.html")
	if err != nil {
		panic(err)
	}
	index := template.Must(template.New("").Parse(string(indexRaw)))

	http.Handle("/img/", http.FileServer(http.FS(assets)))
	http.Handle("/css/", http.FileServer(http.FS(assets)))
	http.Handle("/js/", http.FileServer(http.FS(assets)))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) <= 1 {
			roomID := uuid.NewString()
			http.Redirect(w, r, "/room/"+roomID, http.StatusTemporaryRedirect)
		}
	})

	http.HandleFunc("/room/", func(w http.ResponseWriter, r *http.Request) {
		roomID := r.URL.Path[6:]
		index.Execute(w, roomID)
	})

	http.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		roomID := r.URL.Path[4:]
		websocket.Handler(GetSession(roomID).serve).ServeHTTP(w, r)
	})
	http.ListenAndServe(":8080", nil)
}
