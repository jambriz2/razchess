class RPC {
    constructor(roomID) {
        var jrpc = new simple_jsonrpc();
        var socket = new WebSocket((window.location.protocol == 'https:' ? 'wss:' : 'ws:') + '//' + window.location.host + '/ws/' + roomID);
        socket.onmessage = function(event) {
            jrpc.messageHandler(event.data);
        };
        jrpc.toStream = function(_msg){
            //if (socket.readyState == 3) location.reload(); // closed socket
            socket.send(_msg);
        };
        socket.onerror = function(error) {
            console.error("Error: " + error.message);
        };
        socket.onclose = function(event) {
            console.info('close code : ' + event.code + ' reason: ' + event.reason + ' clean: ' + event.wasClean);
            location.reload();
        };
        this.jrpc = jrpc;
    }

    sendMove(san) {
        var serverResponse = null;
        this.jrpc.call('Session.Move', [san]).then(function(response) {
            serverResponse = response;
        });
        return serverResponse;
    }

    onUpdate(func) {
        this.jrpc.on('Session.Update', function(update) {
            func(update)
            return true
        })
    }
}

class Game {
    constructor(rpc, boardID) {
        this.rpc = rpc;
        this.game = new Chess();
        this.board = Chessboard(boardID, this.getBoardConfig());
        this.wm = null;
        this.bm = null;
        var instance = this;
        rpc.onUpdate(function(update) {
            instance.update(update);
        })
        $(window).resize(function() {
            instance.board.resize();
            instance.colorSpecialSquares();
        });
        $('#' + boardID).on('contextmenu', '.square-55d63', function(e) {
            if (e.button === 2) {
                $(this).toggleClass('highlight-square');
                e.preventDefault();
            }
        })    
    }

    getBoardConfig() {
        var instance = this;
        var config = {
            draggable: true,
            onDragStart: function(source, piece, position, orientation) {
                return instance.onDragStart(source, piece, position, orientation);
            },
            onDrop: function(source, target) {
                return instance.onDrop(source, target);
            }
        }
        return config;
    }

    update(update) {
        this.game = new Chess(update.fen);
        this.board.position(update.fen);
        this.lastMove = update.move;
        this.colorSpecialSquares();
        if (this.onUpdateCallback) {
            this.onUpdateCallback(this.getStatus(), update.fen, update.pgn)
        }
    }

    onUpdate(func) {
        this.onUpdateCallback = func;
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
        var $board = $('#board');
        $board.find('.square-55d63').removeClass('highlight-move').removeClass('highlight-check');
        $board.find('.square-' + this.lastMove[0]).addClass('highlight-move');
        $board.find('.square-' + this.lastMove[1]).addClass('highlight-move');
        if (this.game.in_check()) {
            var color = this.game.turn();
            var king = [].concat(...this.game.board()).find(p => p !== null && p.type === 'k' && p.color === color);
            if (king) {
                $board.find('.square-' + king.square).addClass('highlight-check');
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
        if (this.rpc.sendMove(move.san) == false) return 'snapback';
    }
}

class Menu {
    constructor() {
        $('#menuBtn').click(function() {
            $('#menuBar').toggle();
        });        
    }

    update(status, fen, pgn) {
        var sessionUrl = window.location.protocol + '//' + window.location.host + '/room/' + $('#roomID').val();
        var html = '<span>';
        html += ' <a href="/">New session</a>'
        html += ' | <a href="#" onClick="navigator.clipboard.writeText(\'' + sessionUrl + '\'); return false;">Copy session link</a>';
        html += ' | <a href="#" onClick="navigator.clipboard.writeText(\'' + fen + '\'); return false;">Copy FEN</a>';
        html += ' | <a href="#" onClick="navigator.clipboard.writeText(\'' + pgn + '\'); return false;">Copy PGN</a>';
        html += ' | <a href="/puzzle">Play a puzzle</a>';
        html += '</span>'
        html += '<br />' + status
        $('#menuBar').html(html);
        document.title = status + ' - RazChess'
    }
}

var rpc = new RPC($('#roomID').val());
var menu = new Menu();
rpc.onUpdate(function(update) {
    $('#loading').hide();
    $('#board').show();
    game = new Game(rpc, 'board');
    game.onUpdate(menu.update);
    game.update(update);
})
