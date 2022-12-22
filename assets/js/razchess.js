var room = $('#roomID').val()
var jrpc = new simple_jsonrpc();
var socket = new WebSocket('ws://' + window.location.host + '/ws/' + room);
var config;
var board = null;
var $board = $('#board')
var game = null;
var squareClass = 'square-55d63'

socket.onmessage = function(event) {
    jrpc.messageHandler(event.data);
};

jrpc.toStream = function(_msg){
    socket.send(_msg);
};

socket.onerror = function(error) {
    console.error("Error: " + error.message);
};

socket.onclose = function(event) {
    if (event.wasClean) {
        console.info('Connection close was clean');
    } else {
        console.error('Connection suddenly close');
    }
    console.info('close code : ' + event.code + ' reason: ' + event.reason);
};

jrpc.on('Session.Update', function(update) {
    game = new Chess(update.fen);
    if (!board) {
        $('#loading').hide();
        $('#board').show();
        board = Chessboard('board', config);
    }
    board.position(update.fen);

    $board.find('.' + squareClass).removeClass('highlight-white')
    $board.find('.square-' + update.wm[0]).addClass('highlight-white')
    $board.find('.square-' + update.wm[1]).addClass('highlight-white')

    $board.find('.' + squareClass).removeClass('highlight-black')
    $board.find('.square-' + update.bm[0]).addClass('highlight-black')
    $board.find('.square-' + update.bm[1]).addClass('highlight-black')

    return true;
})

function onDragStart (source, piece, position, orientation) {
    if (game.game_over()) return false;

    if ((game.turn() === 'w' && piece.search(/^b/) !== -1) ||
        (game.turn() === 'b' && piece.search(/^w/) !== -1)) {
        return false;
    }

    /*if (game.turn() === 'w') {
        $board.find('.' + squareClass).removeClass('highlight-white')
    } else {
        $board.find('.' + squareClass).removeClass('highlight-black')
    }*/
}

function onDrop (source, target) {
    var move = game.move({
        from: source,
        to: target,
        promotion: 'q'
    });

    if (move === null) return 'snapback';

    var serverResponse = null;
    jrpc.call('Session.Move', [move.san]).then(function(response) {
        serverResponse = response;
    });
    if (serverResponse == false) return 'snapback';
}

var config = {
  draggable: true,
  onDragStart: onDragStart,
  onDrop: onDrop
};

$(window).resize(function(){
    if (board) board.resize();
});
