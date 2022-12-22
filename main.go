package main

import (
	"bufio"
	"embed"
	"flag"
	"html/template"
	"io/fs"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

//go:embed assets/*
var assets embed.FS

var puzzles []string

func init() {
	rand.Seed(time.Now().Unix())
}

func loadPuzzles() {
	r, err := assets.Open("assets/chess-puzzles.fen")
	if err != nil {
		panic(err)
	}
	defer r.Close()

	puzzles = make([]string, 0, 220)
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		puzzles = append(puzzles, scanner.Text())
	}
}

func main() {
	var mgr SessionMgr
	var addr string
	flag.DurationVar(&mgr.KillTimeout, "session-timeout", defaultKillTimeout, "session expiration time after all players left")
	flag.StringVar(&addr, "addr", ":8080", "http listen address")
	flag.Parse()

	loadPuzzles()

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
			http.Redirect(w, r, "/room/"+uuid.NewString(), http.StatusTemporaryRedirect)
		}
	})

	http.HandleFunc("/room/", func(w http.ResponseWriter, r *http.Request) {
		roomID := r.URL.Path[6:]
		if len(roomID) == 0 {
			http.Redirect(w, r, "/room/"+uuid.NewString(), http.StatusTemporaryRedirect)
		}
		index.Execute(w, roomID)
	})

	http.HandleFunc("/fen/", func(w http.ResponseWriter, r *http.Request) {
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

	http.HandleFunc("/puzzle", func(w http.ResponseWriter, r *http.Request) {
		puzzle := puzzles[rand.Intn(len(puzzles))]
		roomID, err := mgr.NewCustomSession(puzzle)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		} else {
			http.Redirect(w, r, "/room/"+roomID, http.StatusTemporaryRedirect)
		}
	})

	http.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		roomID := r.URL.Path[4:]
		websocket.Handler(mgr.GetSession(roomID).serve).ServeHTTP(w, r)
	})
	http.ListenAndServe(addr, nil)
}
