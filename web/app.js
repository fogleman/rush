new p5();

var MinPieceSize = 2;
var MaxPieceSize = 3;

var BackgroundColor   = "#FFFFFF";
var BoardColor        = "#F2EACD";
var GridLineColor     = "#222222";
var PrimaryPieceColor = "#CC3333";
var PieceColor        = "#338899";
var PieceOutlineColor = "#222222";
var LabelColor        = "#222222";
var WallColor         = "#222222";
var WallBoltColor     = "#AAAAAA";

// Piece

function Piece(position, size, stride) {
    this.position = position;
    this.size = size;
    this.stride = stride;
    this.fixed = size === 1;
}

Piece.prototype.move = function(steps) {
    this.position += this.stride * steps;
}

Piece.prototype.draw = function(boardSize, offset) {
    offset = offset || 0;
    var i0 = this.position;
    var i1 = i0 + this.stride * (this.size - 1);
    var x0 = Math.floor(i0 % boardSize);
    var y0 = Math.floor(i0 / boardSize);
    var x1 = Math.floor(i1 % boardSize);
    var y1 = Math.floor(i1 / boardSize);
    var p = 0.1;
    var x = x0 + p;
    var y = y0 + p;
    var w = x1 - x0 + 1 - p * 2;
    var h = y1 - y0 + 1 - p * 2;
    if (this.stride === 1) {
        x += offset;
    } else {
        y += offset;
    }
    rect(x, y, w, h, 0.1);
}

Piece.prototype.axis = function(point) {
    if (this.stride === 1) {
        return point.x;
    } else {
        return point.y;
    }
}

// Move

function Move(piece, steps) {
    this.piece = piece;
    this.steps = steps;
}

// Board

function Board(desc) {
    this.pieces = [];

    // determine board size
    this.size = Math.floor(Math.sqrt(desc.length));
    this.size2 = this.size * this.size;
    if (this.size2 !== desc.length) {
        throw "boards must be square";
    }

    // parse string
    var positions = new Map();
    for (var i = 0; i < desc.length; i++) {
        var label = desc.charAt(i);
        if (!positions.has(label)) {
            positions.set(label, []);
        }
        positions.get(label).push(i);
    }

    // sort piece labels
    var labels = Array.from(positions.keys());
    labels.sort();

    // add pieces
    for (var label of labels) {
        if (label === '.') {
            continue;
        }
        if (label === 'x') {
            continue;
        }
        var ps = positions.get(label);
        if (ps.length < MinPieceSize) {
            throw "piece size < MinPieceSize";
        }
        if (ps.length > MaxPieceSize) {
            throw "piece size > MaxPieceSize";
        }
        var stride = ps[1] - ps[0];
        if (stride !== 1 && stride !== this.size) {
            throw "invalid piece shape";
        }
        for (var i = 2; i < ps.length; i++) {
            if (ps[i] - ps[i-1] !== stride) {
                throw "invalid piece shape";
            }
        }
        var piece = new Piece(ps[0], ps.length, stride);
        this.addPiece(piece);
    }

    // add walls
    if (positions.has('x')) {
        var ps = positions.get('x');
        for (var p of ps) {
            var piece = new Piece(p, 1, 1);
            this.addPiece(piece);
        }
    }
}

Board.prototype.addPiece = function(piece) {
    this.pieces.push(piece);
}

Board.prototype.doMove = function(move) {
    this.pieces[move.piece].move(move.steps);
}

Board.prototype.pieceAt = function(index) {
    for (var i = 0; i < this.pieces.length; i++) {
        var piece = this.pieces[i];
        var p = piece.position;
        for (var j = 0; j < piece.size; j++) {
            if (p === index) {
                return i;
            }
            p += piece.stride;
        }
    }
    return -1;
}

Board.prototype.isOccupied = function(index) {
    return this.pieceAt(index) >= 0;
}

Board.prototype.moves = function() {
    var moves = [];
    var size = this.size;
    for (var i = 0; i < this.pieces.length; i++) {
        var piece = this.pieces[i];
        if (piece.fixed) {
            continue;
        }
        var reverseSteps;
        var forwardSteps;
        if (piece.stride == 1) {
            var x = Math.floor(piece.position % size);
            reverseSteps = -x;
            forwardSteps = size - piece.size - x;
        } else {
            var y = Math.floor(piece.position / size);
            reverseSteps = -y;
            forwardSteps = size - piece.size - y;
        }
        var idx = piece.position - piece.stride;
        for (var steps = -1; steps >= reverseSteps; steps--) {
            if (this.isOccupied(idx)) {
                break;
            }
            moves.push(new Move(i, steps));
            idx -= piece.stride;
        }
        idx = piece.position + piece.size * piece.stride;
        for (var steps = 1; steps <= forwardSteps; steps++) {
            if (this.isOccupied(idx)) {
                break;
            }
            moves.push(new Move(i, steps));
            idx += piece.stride;
        }
    }
    return moves;
}

// GUI

// var board = new Board("BCDDE.BCF.EGB.FAAGHHHI.G..JIKKLLJMM.");
var board = new Board("IBBx..I..LDDJAAL..J.KEEMFFK..MGGHHHM");

function computeScale(size) {
    var xscale = width / size;
    var yscale = height / size;
    return Math.min(xscale, yscale) * 0.9;
}

function mousePoint() {
    var size = board.size;
    var s = computeScale(size);
    var x = (mouseX - width / 2) / s + size / 2;
    var y = (mouseY - height / 2) / s + size / 2;
    return createVector(x, y);
}

function mousePosition() {
    var p = mousePoint();
    var x = Math.floor(p.x);
    var y = Math.floor(p.y);
    return y * board.size + x;
}

var dragPiece = -1;
var dragAnchor;
var dragDelta;

function mousePressed() {
    dragAnchor = mousePoint();
    dragDelta = createVector(0, 0);
    dragPiece = board.pieceAt(mousePosition());
    if (dragPiece < 0) {
        return;
    }
    var piece = board.pieces[dragPiece];
    if (piece.fixed) {
        dragPiece = -1;
    }
}

function mouseReleased() {
    if (dragPiece < 0) {
        return;
    }
    dragDelta = p5.Vector.sub(mousePoint(), dragAnchor);
    var piece = board.pieces[dragPiece];
    var steps = Math.round(piece.axis(dragDelta));
    for (var move of board.moves()) {
        if (move.piece === dragPiece && move.steps === steps) {
            board.doMove(move);
            break;
        }
    }
    dragPiece = -1;
}

function mouseDragged() {
    if (dragPiece < 0) {
        return;
    }
    dragDelta = p5.Vector.sub(mousePoint(), dragAnchor);
}

function setup() {
    createCanvas(windowWidth, windowHeight);
}

function windowResized() {
    resizeCanvas(windowWidth, windowHeight);
}

function draw() {
    // var moves = board.moves();
    // var move = moves[Math.floor(Math.random() * moves.length)];
    // board.doMove(move);

    background(BackgroundColor);

    var size = board.size;
    var s = computeScale(size);

    resetMatrix();
    translate(width / 2, height / 2);
    scale(s);
    translate(-size / 2, -size / 2);

    // board
    fill(BoardColor);
    stroke(GridLineColor);
    strokeWeight(0.03);
    rect(0, 0, size, size, 0.1);

    // walls
    noStroke();
    ellipseMode(RADIUS);
    for (var piece of board.pieces) {
        if (!piece.fixed) {
            continue;
        }
        var x = Math.floor(piece.position % size);
        var y = Math.floor(piece.position / size);
        fill(WallColor);
        rect(x, y, 1, 1);
        var p = 0.15;
        var r = 0.04;
        fill(WallBoltColor);
        ellipse(x + p, y + p, r);
        ellipse(x + 1 - p, y + p, r);
        ellipse(x + p, y + 1 - p, r);
        ellipse(x + 1 - p, y + 1 - p, r);
    }

    // grid lines
    stroke(GridLineColor);
    strokeWeight(0.015);
    for (var i = 1; i < size; i++) {
        line(i, 0, i, size);
        line(0, i, size, i);
    }

    // pieces
    stroke(PieceOutlineColor);
    strokeWeight(0.03);
    for (var i = 0; i < board.pieces.length; i++) {
        if (i === dragPiece) {
            continue;
        }
        var piece = board.pieces[i];
        if (piece.fixed) {
            continue;
        }
        if (i === 0) {
            fill(PrimaryPieceColor);
        } else {
            fill(PieceColor);
        }
        piece.draw(size);
    }

    // dragging
    if (dragPiece >= 0) {
        var piece = board.pieces[dragPiece];
        var offset = piece.axis(dragDelta);
        if (dragPiece === 0) {
            fill(PrimaryPieceColor);
        } else {
            fill(PieceColor);
        }
        piece.draw(size, offset);
    }
}
