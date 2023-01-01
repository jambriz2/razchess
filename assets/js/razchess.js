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
        this.logic = new Chess();
        this.board = Chessboard(boardID, this.getBoardConfig());
        this.wm = null;
        this.bm = null;
        var instance = this;
        rpc.onUpdate(function(update) {
            instance.update(update);
        })
        $(window).resize(function() {
            instance.board.resize();
            instance.colorLastMoves();
        });
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
        this.logic = new Chess(update.fen);
        this.board.position(update.fen);
        this.wm = update.wm;
        this.bm = update.bm;
        this.colorLastMoves();
        if (this.onUpdateCallback) {
            this.onUpdateCallback(this.getStatus(), update.wm, update.bm, update.fen)
        }
    }

    onUpdate(func) {
        this.onUpdateCallback = func;
    }

    resize() {
        this.board.resize();
        this.colorLastMoves();
    }

    getStatus() {
        var moveColor = (this.logic.turn() === 'w' ? 'White' : 'Black')
        if (this.logic.in_checkmate()) {
          return 'Game over, ' + moveColor + ' is in checkmate'
        }
        else if (this.logic.in_draw()) {
          return 'Game over, drawn position'
        }
        else {
          var status = moveColor + ' to move'
          if (this.logic.in_check()) {
            status += ', ' + moveColor + ' is in check'
          }
          return status
        }
    }
    
    colorLastMoves() {
        const squareClass = 'square-55d63'
        var $board = $('#board');
        $board.find('.' + squareClass).removeClass('highlight-white')
        $board.find('.square-' + this.wm[0]).addClass('highlight-white')
        $board.find('.square-' + this.wm[1]).addClass('highlight-white')
        $board.find('.' + squareClass).removeClass('highlight-black')
        $board.find('.square-' + this.bm[0]).addClass('highlight-black')
        $board.find('.square-' + this.bm[1]).addClass('highlight-black')
    }

    onDragStart(source, piece, position, orientation) {
        if (this.logic.game_over()) return false;
        if ((this.logic.turn() === 'w' && piece.search(/^b/) !== -1) ||
            (this.logic.turn() === 'b' && piece.search(/^w/) !== -1)) {
            return false;
        }
    }
    
    onDrop(source, target) {
        var move = this.logic.move({
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

    update(status, wm, bm, fen) {
        var sessionUrl = window.location.protocol + '//' + window.location.host + '/room/' + $('#roomID').val();
        var html = '<span>';
        html += ' <a href="/">New session</a>'
        html += ' | <a href="#" onClick="navigator.clipboard.writeText(\'' + sessionUrl + '\'); return false;">Copy session link</a>';
        html += ' | <a href="#" onClick="navigator.clipboard.writeText(\'' + fen + '\'); return false;">Copy FEN</a>';
        html += ' | <a href="/fen/' + fen + '" target="_blank">Clone session</a>'
        html += ' | <a href="/puzzle">Play a puzzle</a>';
        html += '</span>'
        html += '<br />' + status
        if (wm && wm[0].length > 0) {
            html += ' | ⚪ ' + wm[0] + '→' + wm[1];
        }
        if (bm && bm[0].length > 0) {
            html += ' | ⚫ ' + bm[0] + '→' + bm[1];
        }
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
