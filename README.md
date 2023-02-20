# RazChess

## Game modes
* Normal chess
* Fischer random (chess960)
* Chess puzzles!
  * 220x built-in mate-in-2-steps puzzles
  * External puzzles can be loaded
* Custom game editor to create your own games

## Other features
* Share the room/session link with as many people as you want, they can all watch or participate
* Auto reconnect
* Download your game as a GIF
* Encyclopaedia of Chess Openings included
* Copy the FEN or PGN of the current game to use it elsewhere
* Optional persistent storage using Redis, so you can continue your sessions even after restarting razchess

## Limitations
* No server side built-in bot (but the standalone bot in [tools/bot/](tools/bot/) can connect to your session)
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
  -logfile string
        Optional path to a log file (still logs to stdout)
  -puzzles string
        Optional location of external puzzles (newline separated list of FEN strings)
  -redis string
        Optional Redis connection string (redis://user:pass@host:port)
  -session-timeout duration
        Session expiration time after all players left (default 1h0m0s)
```
