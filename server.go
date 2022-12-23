package main

import (
	"html/template"
	"io/fs"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type Server struct {
	http.ServeMux
}

func NewServer(assets fs.FS, mgr *SessionMgr, puzzles []string) *Server {
	indexRaw, err := fs.ReadFile(assets, "index.html")
	if err != nil {
		panic(err)
	}
	index := template.Must(template.New("").Parse(string(indexRaw)))
	srv := &Server{}

	srv.Handle("/img/", http.FileServer(http.FS(assets)))
	srv.Handle("/css/", http.FileServer(http.FS(assets)))
	srv.Handle("/js/", http.FileServer(http.FS(assets)))

	srv.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) <= 1 {
			http.Redirect(w, r, "/room/"+uuid.NewString(), http.StatusTemporaryRedirect)
		}
	})

	srv.HandleFunc("/room/", func(w http.ResponseWriter, r *http.Request) {
		roomID := r.URL.Path[6:]
		if len(roomID) == 0 {
			http.Redirect(w, r, "/room/"+uuid.NewString(), http.StatusTemporaryRedirect)
		}
		index.Execute(w, roomID)
	})

	srv.HandleFunc("/fen/", func(w http.ResponseWriter, r *http.Request) {
		customFEN := r.URL.Path[5:]
		if len(customFEN) == 0 {
			http.Redirect(w, r, "/room/"+uuid.NewString(), http.StatusTemporaryRedirect)
		}
		roomID, err := mgr.NewCustomSession(customFEN)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Redirect(w, r, "/room/"+roomID, http.StatusTemporaryRedirect)
		}
	})

	srv.HandleFunc("/puzzle", func(w http.ResponseWriter, r *http.Request) {
		puzzle := puzzles[rand.Intn(len(puzzles))]
		roomID, err := mgr.NewCustomSession(puzzle)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Redirect(w, r, "/room/"+roomID, http.StatusTemporaryRedirect)
		}
	})

	srv.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		roomID := r.URL.Path[4:]
		mgr.GetSessionServer(roomID).ServeHTTP(w, r)
	})

	return srv
}
