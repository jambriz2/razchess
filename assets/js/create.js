var activeColor = new class {
    get() {
        return $('input[name="active-color"]:checked').val();
    }
    set(color) {
        $('input[name="active-color"]').filter('[value=' + color + ']').prop('checked', true);
    }
}

var castlingRights = new class {
    get() {
        var result = '';
        if ($('#white-king-side-castle').is(':checked')) {
            result += 'K';
        }
        if ($('#white-queen-side-castle').is(':checked')) {
            result += 'Q';
        }
        if ($('#black-king-side-castle').is(':checked')) {
            result += 'k';
        }
        if ($('#black-queen-side-castle').is(':checked')) {
            result += 'q';
        }
        return result === '' ? '-' : result;
    }
    set(castlingRights) {
        if (castlingRights.includes('K')) {
            $('#white-king-side-castle').prop('checked', true);
        } else {
            $('#white-king-side-castle').prop('checked', false);
        }
        if (castlingRights.includes('Q')) {
            $('#white-queen-side-castle').prop('checked', true);
        } else {
            $('#white-queen-side-castle').prop('checked', false);
        }
        if (castlingRights.includes('k')) {
            $('#black-king-side-castle').prop('checked', true);
        } else {
            $('#black-king-side-castle').prop('checked', false);
        }
        if (castlingRights.includes('q')) {
            $('#black-queen-side-castle').prop('checked', true);
        } else {
            $('#black-queen-side-castle').prop('checked', false);
        }
    }
}

var enPassantSquare = new class {
    get() {
        return $('#en-passant-square').val();
    }
    set(square) {
        $('#en-passant-square').val(square).change();
    }
}

var fen = new class {
    get() {
        return board.fen() + ' ' + activeColor.get() + ' ' + castlingRights.get() + ' ' + enPassantSquare.get() + ' 0 1';
    }
    set(fen) {
        $('#fen').val(fen);
        updateControls();
    }
}

function updateControls() {
    var fen = $('#fen').val();
    var parts = fen.split(' ');
    if (parts.length != 6) return;
    board.position(fen);
    activeColor.set(parts[1]);
    castlingRights.set(parts[2]);
    enPassantSquare.set(parts[3]);
}

function updateFEN() {
    setTimeout(() => {
        var fenStr = fen.get();
        $('#fen').val(fenStr);
        history.replaceState({fen: fenStr}, null, "/create/" + fenStr.replaceAll(' ', '_'));
    }, 0);
}

function setBoardDefault() {
    board.start();
    castlingRights.set('KQkq');
    activeColor.set('w');
    enPassantSquare.set('-');
}

function setBoardClear() {
    board.clear();
    castlingRights.set('');
    enPassantSquare.set('-');
}

var board = Chessboard('editorBoard', {
    draggable: true,
    dropOffBoard: 'trash',
    sparePieces: true,
    onChange: updateFEN
});
$(window).resize(function () {
    board.resize();
});

fen.set($("#fen").val());
