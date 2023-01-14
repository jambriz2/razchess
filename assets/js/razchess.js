//var loadingSVG = $.ajax({url: '/img/loading.svg', async: false}).responseText;
var loadingSVG = $('#loader').html();

class Game {
    constructor(roomID, boardID) {
        this.roomID = roomID;
        this.boardID = boardID;
        this.$board = $('#' + boardID);
        this.game = new Chess();
        this.setLoading();
        this.connectToRPC();
        var self = this;
        $(window).resize(function() {
            self.board.resize();
            self.colorSpecialSquares();
        });
        $('#' + boardID).on('contextmenu', '.square-55d63', function(e) {
            if (e.button === 2) {
                $(this).toggleClass('highlight-square');
                e.preventDefault();
            }
        })
    }

    connectToRPC() {
        var self = this;
        var jrpc = new simple_jsonrpc();
        var socket = new WebSocket((window.location.protocol == 'https:' ? 'wss:' : 'ws:') + '//' + window.location.host + '/ws/' + roomID);
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
            self.handleDisconnect(error);
        };
        socket.onclose = function(event) {
            console.info('close code : ' + event.code + ' reason: ' + event.reason + ' clean: ' + event.wasClean);
            self.handleDisconnect(event);
        };
        this.jrpc = jrpc;
    }

    setLoading() {
        this.$board.html('<div class="loading">' + loadingSVG + '</div>');
    }

    handleDisconnect() {
        if (this.board) {
            this.board.destroy();
            this.board = null;
            if (this.onUpdate) {
                this.onUpdate('Disconnected', '', this.lastFEN, this.lastPGN);
            }
            this.setLoading();
        }
        var self = this;
        setTimeout(() => { self.connectToRPC(); }, 5000);
    }

    getBoardConfig() {
        var self = this;
        var config = {
            draggable: true,
            onDragStart: function(source, piece, position, orientation) {
                return self.onDragStart(source, piece, position, orientation);
            },
            onDrop: function(source, target) {
                return self.onDrop(source, target);
            }
        }
        return config;
    }

    sendMove(san) {
        var serverResponse = null;
        this.jrpc.call('Session.Move', [san]).then(function(response) {
            serverResponse = response;
        });
        return serverResponse;
    }

    update(update) {
        this.game = new Chess(update.fen);
        if (!this.board) {
            this.board = Chessboard(this.boardID, this.getBoardConfig());
        }
        this.board.position(update.fen);
        this.lastMove = update.move;
        this.lastFEN = update.fen;
        this.lastPGN = update.pgn;
        this.colorSpecialSquares();
        if (this.onUpdate) {
            this.onUpdate(this.getStatus(), update.opening, update.fen, update.pgn)
        }
    }

    resize() {
        this.board.resize();
        this.colorSpecialSquares();
    }

    getStatus() {
        var moveColor = (this.game.turn() === 'w' ? 'White' : 'Black');
        if (this.game.in_checkmate()) {
          return 'Game over, ' + moveColor + ' is in checkmate';
        }
        else if (this.game.in_draw()) {
          return 'Game over, drawn position';
        }
        else {
          var status = moveColor + ' to move';
          if (this.game.in_check()) {
            status += ', ' + moveColor + ' is in check';
          }
          return status;
        }
    }
    
    colorSpecialSquares() {
        this.$board.find('.square-55d63').removeClass('highlight-move').removeClass('highlight-check');
        if (this.lastMove) {
            this.$board.find('.square-' + this.lastMove[0]).addClass('highlight-move');
            this.$board.find('.square-' + this.lastMove[1]).addClass('highlight-move');
        }
        if (this.game.in_check()) {
            var color = this.game.turn();
            var king = [].concat(...this.game.board()).find(p => p !== null && p.type === 'k' && p.color === color);
            if (king) {
                this.$board.find('.square-' + king.square).addClass('highlight-check');
            }
        }
    }

    onDragStart(source, piece, position, orientation) {
        if (this.game.game_over()) return false;
        if ((this.game.turn() === 'w' && piece.search(/^b/) !== -1) ||
            (this.game.turn() === 'b' && piece.search(/^w/) !== -1)) {
            return false;
        }
    }
    
    onDrop(source, target) {
        var move = this.game.move({
            from: source,
            to: target,
            promotion: 'q'
        });
        if (move === null) return 'snapback';
        if (this.sendMove(move.san) == false) return 'snapback';
    }
}

class Menu {
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

    update(status, opening, fen, pgn) {
        this.fen = fen;
        this.pgn = pgn;
        var html = '<span>' + status + '</span>';
        if (opening) {
            html = '<h1>' + opening + '</h1> - ' + status;
        }
        $('#status').html(html);
        document.title = status + ' - RazChess'
    }
}

var roomID = $('#roomID').val();
var menu = new Menu();
var game = new Game(roomID, 'board');
game.onUpdate = function(status, fen, pgn) {
    menu.update(status, fen, pgn);
};
