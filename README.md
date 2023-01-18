# RazChess

## Features
* Completely hassle free, no registration required
* Super easy to create your own session
  * Visit the website and get redirected to your own session
  * Create your own custom room name {razchess-host}/room/{custom-room-name}
* Play Fischer random (chess960) games
* Play chess puzzles!
  * 220x built-in mate-in-2-steps puzzles
  * External puzzles can be loaded
* Create custom games from FEN or PGN strings
* Share the room/session link with as many people as you want, they can all watch or participate
* Download your game as a GIF
* Encyclopaedia of Chess Openings included
* Optional persistent storage using Redis, so you can continue your sessions even after restarting razchess

## Limitations
* **Pawns can only be promoted to queen**
* No timer
* No option to ask for a draw
* No built-in chat
* Not designed to be scalable, though it could work with sticky sessions or DNS load balancing
* No built-in TLS/https handling, but you can use [razvhost](https://github.com/razzie/razvhost) for that

## Usage
```
Usage of razchess:
  -addr string
        Http listen address (default ":8080")
  -puzzles string
        Optional location of external puzzles (newline separated list of FEN strings)
  -redis string
        Optional Redis connection string (redis://user:pass@host:port)
  -session-timeout duration
        Session expiration time after all players left (default 1h0m0s)
```
