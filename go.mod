module github.com/razzie/razchess

go 1.19

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/notnil/chess v1.9.0
	github.com/razzie/blunder v0.0.0-20230219205641-ce2a3d968a7e
	github.com/razzie/chessimage v0.0.0-20230115212848-8c813dc69373
	github.com/razzie/jsonrpc v0.0.0-20230101121601-7e74c3bf4ae5
	golang.org/x/net v0.4.0
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/fogleman/gg v1.1.0 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	golang.org/x/exp v0.0.0-20220518171630-0b5c67f07fdf // indirect
	golang.org/x/image v0.3.0 // indirect
	gopkg.in/freeeve/pgn.v1 v1.0.1 // indirect
)

// go list -f '{{.Version}}' -m github.com/razzie/chess@master
replace github.com/notnil/chess => github.com/razzie/chess v1.9.1-0.20230216225629-5022223cc522
