package main

import (
	"bufio"
	"embed"
	"flag"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/razzie/razchess/pkg/razchess"
)

//go:embed assets/*
var assets embed.FS

func loadPuzzles(filename string) (puzzles []string) {
	if len(filename) == 0 {
		return
	}
	r, err := os.Open(filename)
	if err != nil {
		log.Println("failed to load puzzles:", err)
		return
	}
	defer r.Close()
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		puzzles = append(puzzles, scanner.Text())
	}
	return
}

func main() {
	var redisURL string
	var killTimeout time.Duration
	var puzzlesFilename string
	var logfile string
	flag.StringVar(&redisURL, "redis", "", "Optional Redis connection string (redis://user:pass@host:port)")
	flag.DurationVar(&killTimeout, "session-timeout", razchess.DefaultKillTimeout, "Session expiration time after all players left")
	flag.StringVar(&puzzlesFilename, "puzzles", "", "Optional location of external puzzles (newline separated list of FEN strings)")
	flag.StringVar(&logfile, "logfile", "", "Optional path to a log file (still logs to stdout)")
	flag.Parse()

	if len(logfile) > 0 {
		f, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		if err != nil {
			log.SetOutput(os.Stdout)
			log.Println(err)
		} else {
			defer f.Close()
			mw := io.MultiWriter(os.Stdout, f)
			log.SetOutput(mw)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("$PORT must be set")
	}
	addr := ":" + port

	assets, _ := fs.Sub(assets, "assets")
	mgr := razchess.NewSessionMgr(redisURL, killTimeout)
	srv := razchess.NewServer(assets, mgr, loadPuzzles(puzzlesFilename))

	log.Println("[RazChess server started]")
	http.ListenAndServe(addr, srv)
}

