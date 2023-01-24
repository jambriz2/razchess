var loadingSVG = $('#loader').html();
var sounds = {
    move: new Audio('/sounds/move.ogg'),
    capture: new Audio('/sounds/capture.ogg'),
    check: new Audio('/sounds/check.ogg'),
    illegal: new Audio('/sounds/illegal.ogg'),
    gameOver: new Audio('/sounds/game-over.ogg')
};

class Game {
    #roomID;
    #boardID;
    #board;
    #$board;
    #orientation;
    #state;
    #jrpc;
    onUpdate;

    constructor(roomID, boardID) {
        this.#roomID = roomID;
        this.#boardID = boardID;
        this.#$board = $('#' + boardID);
        this.#setLoading();
        this.#connectToRPC();
        var self = this;
        $(window).resize(function() {
            self.#resize();
        });
    }

    #connectToRPC() {
        var self = this;
        var jrpc = new simple_jsonrpc();
        var socket = new WebSocket((window.location.protocol == 'https:' ? 'wss:' : 'ws:') + '//' + window.location.host + '/ws/' + this.#roomID);
        jrpc.on('Session.Update', function(update) {
            self.update(update);
            return true;
        })
        jrpc.toStream = function(_msg){
            socket.send(_msg);
        };
        socket.onmessage = function(event) {
            jrpc.messageHandler(event.data);
        };
        socket.onerror = function(error) {
            console.error("Error: " + error.message);
            self.#handleDisconnect(error);
        };
        socket.onclose = function(event) {
            console.info('close code : ' + event.code + ' reason: ' + event.reason + ' clean: ' + event.wasClean);
            self.#handleDisconnect(event);
        };
        this.#jrpc = jrpc;
    }

    #setLoading() {
        this.#$board.html('<div class="loading">' + loadingSVG + '</div>');
    }

    #handleDisconnect() {
        if (this.#board) {
            this.#board.destroy();
            this.#board = null;
            if (this.onUpdate) {
                this.onUpdate({
                    status: 'Disconnected',
                    fen: this.#state.fen,
                    pgn: this.#state.pgn
                });
            }
            this.#setLoading();
        }
        var self = this;
        setTimeout(() => { self.#connectToRPC(); }, 1000);
    }

    sendMove(move) {
        return this.#jrpc.call('Session.Move', [move]);
    }

    resign() {
        return this.#jrpc.call('Session.Resign', [this.#state.turn]);
    }

    update(update) {
        if (!this.#board) {
            this.#createBoard();
        } else {
            this.#playSound(update);
        }
        this.#board.position(update.fen);
        this.#state = update;
        this.#colorSpecialSquares();
        if (this.onUpdate) {
            this.onUpdate(update);
        }
    }

    #playSound(update) {
        if (update.isGameOver) {
            sounds.gameOver.play();
        } else if (update.isCapture) {
            sounds.capture.play();
        } else if (update.checkedSquare) {
            sounds.check.play();
        } else if (update.move) {
            sounds.move.play();
        }
    }

    #createBoard() {
        var self = this;
        var config = {
            draggable: true,
            onDragStart: function(source, piece, position, orientation) {
                return self.#onDragStart(source, piece, position, orientation);
            },
            onDrop: function(source, target, piece) {
                return self.#onDrop(source, target, piece);
            }
        }
        this.#board = Chessboard(this.#boardID, config);
        this.#board.orientation(this.#orientation);
        this.#$board.on('contextmenu', '.square-55d63', function(e) {
            $(this).toggleClass('highlight-square');
            e.preventDefault();
        });
    }

    #resize() {
        if (this.#board) {
            this.#board.resize();
            this.#colorSpecialSquares();
        }
    }

    flipBoard() {
        if (this.#board) {
            this.#board.flip();
            this.#orientation = this.#board.orientation();
            this.#colorSpecialSquares();
        }
    }
    
    #colorSpecialSquares() {
        this.#$board.find('.square-55d63').removeClass('highlight-move').removeClass('highlight-check');
        if (this.#state.move) {
            this.#$board.find('.square-' + this.#state.move[0]).addClass('highlight-move');
            this.#$board.find('.square-' + this.#state.move[1]).addClass('highlight-move');
        }
        if (this.#state.checkedSquare) {
            this.#$board.find('.square-' + this.#state.checkedSquare).addClass('highlight-check');
        }
    }

    #onDragStart(source, piece, position, orientation) {
        if (this.#state.isGameOver) return false;
        if ((this.#state.turn === 'White' && piece.search(/^b/) !== -1) ||
            (this.#state.turn === 'Black' && piece.search(/^w/) !== -1)) {
            return false;
        }
    }
    
    #onDrop(source, target, piece) {
        var game = this;
        var move = source + target;
        if ((piece === 'wP' && target.charAt(1) === '8') || (piece === 'bP' && target.charAt(1) === '1')) {
            move += 'q';
        }
        this.sendMove(move).then(function(valid) {
            if (!valid) {
                game.#board.position(game.#state.fen);
                sounds.illegal.play();
            }
        });
    }
}

var menu = new class {
    copySessionLink() {
        var sessionUrl = window.location.protocol + '//' + window.location.host + '/room/' + $('#roomID').val();
        navigator.clipboard.writeText(sessionUrl);
    }

    copyFEN() {
        navigator.clipboard.writeText(this.fen);
    }

    copyPGN() {
        navigator.clipboard.writeText(this.pgn);
    }

    update(update) {
        this.fen = update.fen;
        this.pgn = update.pgn;
        var html = '<span>' + update.status + '</span>';
        if (update.opening) {
            html = '<h1>' + update.opening + '</h1> - ' + html;
        }
        $('#status').html(html);
        document.title = update.status + ' - RazChess'
    }

    createCustomGame() {
        if (this.fen) {
            var game = this.fen.replaceAll(' ', '_');
            window.location.href = '/create/' + game;
        } else {
            window.location.href = '/create';
        }
    }
}

var roomID = $('#roomID').val();
var game = new Game(roomID, 'board');
game.onUpdate = function(status, fen, pgn) {
    menu.update(status, fen, pgn);
};
