package razchess

import (
	"fmt"
	"html/template"
	"io/fs"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func init() {
	rand.Seed(time.Now().Unix())
}

type Server struct {
	http.ServeMux
	mgr   *SessionMgr
	index *template.Template
}

func NewServer(assets fs.FS, mgr *SessionMgr, puzzles []string) *Server {
	indexRaw, err := fs.ReadFile(assets, "index.html")
	if err != nil {
		panic(err)
	}
	srv := &Server{
		mgr:   mgr,
		index: template.Must(template.New("").Parse(string(indexRaw))),
	}

	if len(puzzles) == 0 {
		puzzles = GetInternalPuzzles()
	}

	srv.Handle("/img/", http.FileServer(http.FS(assets)))
	srv.Handle("/css/", http.FileServer(http.FS(assets)))
	srv.Handle("/js/", http.FileServer(http.FS(assets)))

	srv.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) <= 1 {
			srv.redirectToNewSession(w, r)
		}
	})

	srv.HandleFunc("/room/", func(w http.ResponseWriter, r *http.Request) {
		roomID := r.URL.Path[6:]
		if len(roomID) == 0 {
			srv.redirectToNewSession(w, r)
		}
		srv.index.Execute(w, roomID)
	})

	srv.HandleFunc("/custom/", func(w http.ResponseWriter, r *http.Request) {
		game := r.URL.Path[8:]
		srv.handleCustomSession(w, r, game, true)
	})

	srv.HandleFunc("/puzzle", func(w http.ResponseWriter, r *http.Request) {
		puzzleID := rand.Intn(len(puzzles))
		http.Redirect(w, r, "/puzzle/"+fmt.Sprint(puzzleID), http.StatusTemporaryRedirect)
	})

	srv.HandleFunc("/puzzle/", func(w http.ResponseWriter, r *http.Request) {
		puzzleID, err := strconv.Atoi(r.URL.Path[8:])
		if err != nil || puzzleID < 0 || puzzleID >= len(puzzles) {
			http.Redirect(w, r, "/puzzle", http.StatusTemporaryRedirect)
		}
		srv.handleCustomSession(w, r, puzzles[puzzleID], false)
	})

	srv.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		roomID := r.URL.Path[4:]
		mgr.ServeSession(w, r, roomID)
	})

	return srv
}

func (srv *Server) handleCustomSession(w http.ResponseWriter, r *http.Request, game string, showRoomID bool) {
	if len(game) == 0 {
		srv.redirectToNewSession(w, r)
	}
	roomID, err := srv.mgr.CreateSession(game)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else if showRoomID {
		http.Redirect(w, r, "/room/"+roomID, http.StatusTemporaryRedirect)
	} else {
		srv.index.Execute(w, roomID)
	}
}

func (srv *Server) redirectToNewSession(w http.ResponseWriter, r *http.Request) {
	roomID, _ := srv.mgr.CreateSession("")
	http.Redirect(w, r, "/room/"+roomID, http.StatusTemporaryRedirect)
}
