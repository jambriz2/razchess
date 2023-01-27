var loadingSVG = $('#loader').html();
var sounds = {
    move: new Audio('/sounds/move.ogg'),
    capture: new Audio('/sounds/capture.ogg'),
    check: new Audio('/sounds/check.ogg'),
    illegal: new Audio('/sounds/illegal.ogg'),
    gameOver: new Audio('/sounds/game-over.ogg')
};

function stopRightClick(e) {
    if (e.button != 0) {
        e.stopPropagation();
        e.stopImmediatePropagation();
        e.preventDefault();
        return false;
    }
}

$("body")
    .on('contextmenu', '.piece-417db', stopRightClick)
    .on('mouseup', '.piece-417db', stopRightClick)
    .on('mousedown', '.piece-417db', stopRightClick);

class Game {
    #roomID;
    #boardID;
    #board;
    #$board;
    #orientation;
    #state;
    #jrpc;
    onUpdate;
    onPromotion;

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
            socket.close();
            self.#handleDisconnect(error);
        };
        socket.onclose = function(event) {
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

    move(move) {
        var self = this;
        this.#jrpc.call('Session.Move', [move]).then(function(valid) {
            if (!valid && self.#board) {
                self.#board.position(self.#state.fen);
                sounds.illegal.play();
            }
        });
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
        } else if (update.move[0]) {
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
            onDrop: function(source, target, piece, newPos, oldPos) {
                return self.#onDrop(source, target, piece, newPos, oldPos);
            }
        }
        this.#board = Chessboard(this.#boardID, config);
        this.#board.orientation(this.#orientation);
        this.#$board.on('contextmenu', '.square-55d63', function(e) {
            $(this).toggleClass('highlight-square');
            e.preventDefault();
        });
        this.#$board.find(".square-55d63").on('mousedown', '.piece-417db', stopRightClick);
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
        if (this.#state.move[0]) {
            this.#$board.find('.square-' + this.#state.move[0]).addClass('highlight-move');
            this.#$board.find('.square-' + this.#state.move[1]).addClass('highlight-move');
        }
        if (this.#state.checkedSquare) {
            this.#$board.find('.square-' + this.#state.checkedSquare).addClass('highlight-check');
        }
    }

    #isPawnPromotion(source, target, board) {
        var piece = board[source];
        if (piece != 'wP' && piece != 'bP') {
            return false;
        }
        var color = piece.charAt(0);
        var sourceRank = source.charAt(1);
        var targetRank = target.charAt(1);
        if (color === 'w' && (sourceRank != '7' || targetRank != '8')) {
            return false;
        }
        if (color === 'b' && (sourceRank != '2' || targetRank != '1')) {
            return false;
        }
        var sourceFile = source.charCodeAt(0);
        var targetFile = target.charCodeAt(0);
        if (targetFile === sourceFile && !board[target]) return true;
        return (targetFile == sourceFile+1 || targetFile == sourceFile-1) && board[target];
    }

    #tryHandlePawnPromotionDlg(move) {
        if (!this.onPromotion) return false;
        var self = this;
        this.onPromotion(this.#state.turn).then(
            function(promoteTo) { // resolved
                self.move(move + promoteTo);
            },
            function() { // rejected
                if (self.#board) {
                    self.#board.position(self.#state.fen);
                }
            });
        return true;
    }

    #onDragStart(source, piece, position, orientation) {
        if (this.#state.isGameOver) return false;
        if ((this.#state.turn === 'w' && piece.search(/^b/) !== -1) ||
            (this.#state.turn === 'b' && piece.search(/^w/) !== -1)) {
            return false;
        }
    }
    
    #onDrop(source, target, piece, newPos, oldPos) {
        if (source === target) return;
        var move = source + target;
        if (this.#isPawnPromotion(source, target, oldPos)) {
            if (this.#tryHandlePawnPromotionDlg(move)) return;
            move += 'q';
        }
        this.move(move);
    }
}

class Menu {
    #sessionURL;
    #$status;
    #fen;
    #pgn;

    constructor(roomID, statusDivID) {
        this.#sessionURL = window.location.protocol + '//' + window.location.host + '/room/' + roomID;
        this.#$status = $('#' + statusDivID);
    }

    copySessionLink() {
        navigator.clipboard.writeText(this.#sessionURL);
    }

    copyFEN() {
        navigator.clipboard.writeText(this.#fen);
    }

    copyPGN() {
        navigator.clipboard.writeText(this.#pgn);
    }

    update(update) {
        this.#fen = update.fen;
        this.#pgn = update.pgn;
        var html = '<span>' + update.status + '</span>';
        if (update.opening) {
            html = '<h1>' + update.opening + '</h1> - ' + html;
        }
        this.#$status.html(html);
    }

    createCustomGame() {
        if (this.#fen) {
            var game = this.#fen.replaceAll(' ', '_');
            window.location.href = '/create/' + game;
        } else {
            window.location.href = '/create';
        }
    }
}

class PawnPromotion {
    #$dialog;
    #$dlgImages;
    #piecesTheme;
    #$parent;
    #promotionResolve;
    #promotionReject;

    constructor(dialogID, parentID) {
        this.#$dialog = $('#' + dialogID);
        this.#$dlgImages = $('#' + dialogID + ' img');
        this.#piecesTheme = '/img/chesspieces/wikipedia/';
        this.#$parent = $('#' + parentID);
    }

    openPromotionDlg(color) {
        this.#reject("new dialog opened");
        this.#$dlgImages.filter('[data-piece="q"]').attr('src', this.#piecesTheme + color + 'Q.png');
        this.#$dlgImages.filter('[data-piece="r"]').attr('src', this.#piecesTheme + color + 'R.png');
        this.#$dlgImages.filter('[data-piece="n"]').attr('src', this.#piecesTheme + color + 'N.png');
        this.#$dlgImages.filter('[data-piece="b"]').attr('src', this.#piecesTheme + color + 'B.png');
        var self = this;
        this.#$dialog.dialog({
            modal: true,
            width: this.#$parent.width()/2,
            resizable: false,
            draggable: false,
            closeOnEscape: true,
            close: function() {
                self.#reject("dialog closed");
            },
        }).dialog('widget').position({
            of: this.#$parent,
            my: 'middle middle',
            at: 'middle middle',
        })
        return new Promise((resolve, reject) => {
            this.#promotionResolve = resolve;
            this.#promotionReject = reject;
        });
    }

    promoteTo(piece) {
        if (this.#promotionResolve) {
            this.#promotionResolve(piece);
            this.#promotionResolve = null;
        }
        this.#$dialog.dialog('close');
    }

    #reject(reason) {
        if (this.#promotionReject) {
            this.#promotionReject(reason);
            this.#promotionReject = null;
        }
    }
}

var roomID = $('#roomID').val();
var menu = new Menu(roomID, 'status');
var promotion = new PawnPromotion('promotion-dialog', 'board');
var game = new Game(roomID, 'board');
game.onUpdate = function(update) {
    menu.update(update);
    document.title = update.status + ' - RazChess'
};
game.onPromotion = function(color) {
    return promotion.openPromotionDlg(color);
};
