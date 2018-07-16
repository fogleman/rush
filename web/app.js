// Constants

var UnusableHeight = 32 + 24 * 3;

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

Piece.prototype.draw = function(p5, boardSize, offset) {
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
    p5.rect(x, y, w, h, 0.1);
}

Piece.prototype.pickAxis = function(point) {
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
    if (this.size === 0) {
        throw "board cannot be empty";
    }

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
        if (label === '.' || label === 'o') {
            continue;
        }
        if (label === 'x') {
            continue;
        }
        var ps = positions.get(label);
        if (ps.length < 2) {
            throw "piece size must be >= 2";
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

    // compute some stuff
    this.primaryRow = 0;
    if (this.pieces.length !== 0) {
        this.primaryRow = Math.floor(this.pieces[0].position / this.size);
    }
}

Board.prototype.addPiece = function(piece) {
    this.pieces.push(piece);
}

Board.prototype.doMove = function(move) {
    this.pieces[move.piece].move(move.steps);
}

Board.prototype.undoMove = function(move) {
    this.pieces[move.piece].move(-move.steps);
}

Board.prototype.isSolved = function() {
    if (this.pieces.length === 0) {
        return false;
    }
    var piece = this.pieces[0];
    var x = Math.floor(piece.position % this.size);
    return x + piece.size === this.size;
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

// View

function View() {
    this.board = new Board(".");
    this.dragPiece = -1;
    this.dragAnchor = null;
    this.dragDelta = null;
    this.dragMin = 0;
    this.dragMax = 0;
    this.undoStack = [];

    this.backgroundColor   = "#FFFFFF";
    this.boardColor        = "#F2EACD";
    this.gridLineColor     = "#222222";
    this.primaryPieceColor = "#CC3333";
    this.pieceColor        = "#338899";
    this.pieceOutlineColor = "#222222";
    this.wallColor         = "#222222";
    this.wallBoltColor     = "#AAAAAA";
}

View.prototype.bind = function(p5) {
    this.p5 = p5;
}

View.prototype.setBoard = function(board) {
    this.board = board;
    this.undoStack = [];
    this.changed();
}

View.prototype.computeScale = function() {
    var p5 = this.p5;
    var board = this.board;
    var xscale = (p5.width / board.size) * 0.9;
    var yscale = (p5.height / board.size) * 0.99;
    return Math.min(xscale, yscale);
};

View.prototype.mouseVector = function() {
    var p5 = this.p5;
    var board = this.board;
    var mx = p5.mouseX || p5.touchX;
    var my = p5.mouseY || p5.touchY;
    var scale = this.computeScale();
    var x = (mx - p5.width / 2) / scale + board.size / 2;
    var y = (my - p5.height / 2) / scale + board.size / 2;
    return p5.createVector(x, y);
};

View.prototype.mouseIndex = function() {
    var p5 = this.p5;
    var board = this.board;
    var p = this.mouseVector();
    var x = Math.floor(p.x);
    var y = Math.floor(p.y);
    return y * board.size + x;
};

View.prototype.mousePressed = function() {
    var p5 = this.p5;
    var board = this.board;
    this.dragAnchor = this.mouseVector();
    this.dragDelta = p5.createVector(0, 0);
    this.dragPiece = board.pieceAt(this.mouseIndex());
    if (this.dragPiece < 0) {
        return;
    }
    var piece = board.pieces[this.dragPiece];
    // can't move walls
    if (piece.fixed) {
        this.dragPiece = -1;
        return;
    }
    // determine max range
    this.dragMin = 0;
    this.dragMax = 0;
    for (var move of board.moves()) {
        if (move.piece === this.dragPiece) {
            this.dragMin = Math.min(this.dragMin, move.steps);
            this.dragMax = Math.max(this.dragMax, move.steps);
        }
    }
};

View.prototype.mouseReleased = function() {
    var p5 = this.p5;
    var board = this.board;
    if (this.dragPiece < 0) {
        return;
    }
    this.dragDelta = p5.Vector.sub(this.mouseVector(), this.dragAnchor);
    var piece = board.pieces[this.dragPiece];
    var steps = Math.round(piece.pickAxis(this.dragDelta));
    steps = Math.min(steps, this.dragMax);
    steps = Math.max(steps, this.dragMin);
    for (var move of board.moves()) {
        if (move.piece === this.dragPiece && move.steps === steps) {
            board.doMove(move);
            this.undoStack.push(move);
            this.changed();
            break;
        }
    }
    this.dragPiece = -1;
};

View.prototype.mouseDragged = function() {
    var p5 = this.p5;
    if (this.dragPiece < 0) {
        return;
    }
    this.dragDelta = p5.Vector.sub(this.mouseVector(), this.dragAnchor);
};

View.prototype.touchStarted = function() {
    this.mousePressed();
    return false;
};

View.prototype.touchEnded = function() {
    this.mouseReleased();
    return false;
};

View.prototype.touchMoved = function() {
    this.mouseDragged();
    return false;
};

View.prototype.keyPressed = function() {
    var p5 = this.p5;
    if (p5.key === 'U') {
        this.undo();
    } else if (p5.key === 'R') {
        this.reset();
    }
};

View.prototype.reset = function() {
    var board = this.board;
    while (this.undoStack.length > 0) {
        var move = this.undoStack.pop();
        board.undoMove(move);
    }
    this.changed();
};

View.prototype.undo = function() {
    var board = this.board;
    if (this.undoStack.length > 0) {
        var move = this.undoStack.pop();
        board.undoMove(move);
    }
    this.changed();
};

View.prototype.changed = function() {
    var el = document.getElementById('numMoves');
    if (el) {
        el.textContent = this.undoStack.length;
    }
}

View.prototype.setup = function() {
    var p5 = this.p5;
    p5.createCanvas(p5.windowWidth, p5.windowHeight - UnusableHeight);
};

View.prototype.windowResized = function() {
    var p5 = this.p5;
    p5.resizeCanvas(p5.windowWidth, p5.windowHeight - UnusableHeight);
};

View.prototype.draw = function() {
    var p5 = this.p5;
    var board = this.board;
    var size = board.size;

    p5.background(this.backgroundColor);
    p5.strokeJoin(p5.ROUND);

    var scale = this.computeScale();
    p5.resetMatrix();
    p5.translate(p5.width / 2, p5.height / 2);
    p5.scale(scale);
    p5.translate(-size / 2, -size / 2);

    // exit
    var ex = size;
    var ey = board.primaryRow + 0.5;
    var es = 0.1;
    p5.fill(this.gridLineColor);
    p5.noStroke();
    p5.beginShape();
    p5.vertex(ex, ey + es);
    p5.vertex(ex, ey - es);
    p5.vertex(ex + es, ey);
    p5.endShape(p5.CLOSE);

    // board
    p5.fill(this.boardColor);
    if (board.isSolved()) {
        if (Date.now() % 500 < 250) {
            p5.fill("#FFFFFF");
        }
    }
    p5.stroke(this.gridLineColor);
    p5.strokeWeight(0.03);
    p5.rect(0, 0, size, size, 0.03);

    // walls
    p5.noStroke();
    p5.ellipseMode(p5.RADIUS);
    for (var piece of board.pieces) {
        if (!piece.fixed) {
            continue;
        }
        var x = Math.floor(piece.position % size);
        var y = Math.floor(piece.position / size);
        p5.fill(this.wallColor);
        p5.rect(x, y, 1, 1);
        var p = 0.15;
        var r = 0.04;
        p5.fill(this.wallBoltColor);
        p5.ellipse(x + p, y + p, r);
        p5.ellipse(x + 1 - p, y + p, r);
        p5.ellipse(x + p, y + 1 - p, r);
        p5.ellipse(x + 1 - p, y + 1 - p, r);
    }

    // grid lines
    p5.stroke(this.gridLineColor);
    p5.strokeWeight(0.015);
    for (var i = 1; i < size; i++) {
        p5.line(i, 0, i, size);
        p5.line(0, i, size, i);
    }

    // pieces
    p5.stroke(this.pieceOutlineColor);
    p5.strokeWeight(0.03);
    for (var i = 0; i < board.pieces.length; i++) {
        if (i === this.dragPiece) {
            continue;
        }
        var piece = board.pieces[i];
        if (piece.fixed) {
            continue;
        }
        if (i === 0) {
            p5.fill(this.primaryPieceColor);
        } else {
            p5.fill(this.pieceColor);
        }
        piece.draw(p5, size);
    }

    // dragging
    if (this.dragPiece >= 0) {
        var piece = board.pieces[this.dragPiece];
        var offset = piece.pickAxis(this.dragDelta);
        offset = Math.min(offset, this.dragMax);
        offset = Math.max(offset, this.dragMin);
        var steps = Math.round(offset);

        // if (this.dragPiece === 0) {
        //     p5.fill(this.primaryPieceColor + "66");
        // } else {
        //     p5.fill(this.pieceColor + "66");
        // }
        // p5.stroke(this.pieceOutlineColor + "66");
        // piece.draw(p5, size, steps);

        if (this.dragPiece === 0) {
            p5.fill(this.primaryPieceColor);
        } else {
            p5.fill(this.pieceColor);
        }
        p5.stroke(this.pieceOutlineColor);
        piece.draw(p5, size, offset);
    }
};

//

function hashToBoard() {
    try {
        var desc = location.hash.substring(1);
        return new Board(desc);
    }
    catch (e) {
        return new Board("IBBx..I..LDDJAAL..J.KEEMFFK..MGGHHHM");
    }
}

var view = new View();

view.setBoard(hashToBoard());

var sketch = function(p) {
    p.Vector = p5.Vector;
    view.bind(p);
    p.draw = function() { view.draw(); }
    p.keyPressed = function() { view.keyPressed(); }
    p.mouseDragged = function() { view.mouseDragged(); }
    p.mousePressed = function() { view.mousePressed(); }
    p.mouseReleased = function() { view.mouseReleased(); }
    p.setup = function() { view.setup(); };
    p.touchEnded = function() { view.touchEnded(); }
    p.touchMoved = function() { view.touchMoved(); }
    p.touchStarted = function() { view.touchStarted(); }
    p.windowResized = function() { view.windowResized(); }
};

new p5(sketch, 'view');

window.onload = function() {
    document.ontouchmove = function(event) {
        event.preventDefault();
    }

    window.onhashchange = function() {
        view.setBoard(hashToBoard());
    }

    document.getElementById('resetButton').onclick = function() {
        view.reset();
    }

    document.getElementById('undoButton').onclick = function() {
        view.undo();
    }
};
