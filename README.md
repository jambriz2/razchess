# RazChess

## Features
* Completely hassle free, no registration required
* Super easy to create your own session
  * Visit the website and get redirectede to your own session
  * Create your own custom room name {razchess-host}/room/{custom-room-name}
* Share the room/session link with as many people as you want, they can all watch or participate
* Last steps are shown for both colors
* Info bar can be enabled by clicking on the chess icon in the top left corner
* Chess puzzles! (220x mate-in-2-steps puzzles as of writing this doc)
* Optional persistent storage using Redis, so you can continue your sessions even after restarting razchess

## Limitations
* No built-in chat
* No timer
* No step count
* No option to resign or ask for a draw
* **Pawns can only be promoted to queen**
* Not designed to be scalable, though it could work with sticky sessions or DNS load balancing
* No built-in TLS/https handling, but you can use [razvhost](https://github.com/razzie/razvhost) for that

## Usage
```
Usage of razchess:
  -addr string
        http listen address (default ":8080")
  -redis string
        Redis connection string (redis://user:pass@host:port)
  -session-timeout duration
        session expiration time after all players left (default 1h0m0s)
```
