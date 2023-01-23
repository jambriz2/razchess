package razchess

import (
	"fmt"
	"html/template"
	"io/fs"
	"math/rand"
	"net/http"
	"net/url"
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
	srv.Handle("/sounds/", http.FileServer(http.FS(assets)))

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

	srv.HandleFunc("/custom", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		game, err := gameFromForm(r.Form)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		srv.serveSession(w, r, game, true)
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
		srv.serveSession(w, r, puzzles[puzzleID], false)
	})

	srv.HandleFunc("/fischer-random", func(w http.ResponseWriter, r *http.Request) {
		srv.serveSession(w, r, GenerateFischerRandomFEN(), true)
	})

	srv.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		roomID := r.URL.Path[4:]
		mgr.ServeRPC(w, r, roomID)
	})

	srv.HandleFunc("/gif/", func(w http.ResponseWriter, r *http.Request) {
		roomID := r.URL.Path[5:]
		w.Header().Set("Content-Disposition", "attachment; filename="+roomID+".gif")
		w.Header().Set("Content-Type", "image/gif")
		if err := mgr.MoveHistoryToGIF(w, roomID); err != nil {
			http.Error(w, "Session not found", http.StatusNotFound)
		}
	})

	return srv
}

func (srv *Server) serveSession(w http.ResponseWriter, r *http.Request, game string, showRoomID bool) {
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
	srv.serveSession(w, r, "", true)
}

func gameFromForm(form url.Values) (string, error) {
	for _, gameType := range []string{"fen", "pgn"} {
		if form.Has(gameType) {
			return gameType + ":" + form.Get(gameType), nil
		}
	}
	return "", fmt.Errorf("invalid form")
}
