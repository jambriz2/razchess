package main

import (
	"bufio"
	"embed"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"time"
)

//go:embed assets/*
var assets embed.FS

func loadPuzzles() []string {
	r, err := assets.Open("assets/chess-puzzles.fen")
	if err != nil {
		panic(err)
	}
	defer r.Close()

	puzzles := make([]string, 0, 220)
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		puzzles = append(puzzles, scanner.Text())
	}
	return puzzles
}

func main() {
	var redisURL string
	var killTimeout time.Duration
	var addr string
	flag.StringVar(&redisURL, "redis", "", "Redis connection string (redis://user:pass@host:port)")
	flag.DurationVar(&killTimeout, "session-timeout", defaultKillTimeout, "session expiration time after all players left")
	flag.StringVar(&addr, "addr", ":8080", "http listen address")
	flag.Parse()

	assets, _ := fs.Sub(assets, "assets")
	mgr := NewSessionMgr(redisURL, killTimeout)
	srv := NewServer(assets, mgr, loadPuzzles())

	log.Println("[RazChess server started]")
	http.ListenAndServe(addr, srv)
}
