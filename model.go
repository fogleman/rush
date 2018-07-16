package rush

import (
	"fmt"
	"image"
	"math"
	"math/rand"
	"sort"
	"strings"
)

// Orientation indicates which direction a Piece can move. Horizontal pieces
// can move left and right. Vertical pieces can move up and down.
type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

// Piece represents a piece (a car or a truck) on the grid. Its position is
// a zero-indexed int, 0 <= Position < W*H. Its size specifies how many cells
// it occupies. Its orientation specifies whether it is vertical or horizontal.
type Piece struct {
	Position    int
	Size        int
	Orientation Orientation
}

func (piece *Piece) Stride(w int) int {
	if piece.Orientation == Horizontal {
		return 1
	}
	return w
}

func (piece *Piece) Row(w int) int {
	return piece.Position / w
}

func (piece *Piece) Col(w int) int {
	return piece.Position % w
}

// Move represents a move to make on the board. Piece indicates which piece
// (by index) to move and Steps is a non-zero positive or negative int that
// specifies how many cells to move the piece.
type Move struct {
	Piece int
	Steps int
}

func (move Move) AbsSteps() int {
	if move.Steps < 0 {
		return -move.Steps
	}
	return move.Steps
}

func (move Move) Label() string {
	return string('A' + move.Piece)
}

func (move Move) String() string {
	return fmt.Sprintf("%s%+d", move.Label(), move.Steps)
}

// Board represents the complete puzzle state. The size of the grid, the
// placement, size, orientation of the pieces. The placement of walls
// (immovable obstacles). Which cells are occupied, either by a piece or a
// wall.
type Board struct {
	Width    int
	Height   int
	Pieces   []Piece
	Walls    []int
	occupied []bool
	memoKey  MemoKey
}

func NewEmptyBoard(w, h int) *Board {
	occupied := make([]bool, w*h)
	memoKey := MakeMemoKey(nil)
	return &Board{w, h, nil, nil, occupied, memoKey}
}

func NewRandomBoard(w, h, primaryRow, primarySize, numPieces, numWalls int) *Board {
	board := NewEmptyBoard(w, h)
	board.AddPiece(Piece{primaryRow * w, primarySize, Horizontal})
	for i := 1; i < numPieces; i++ {
		board.mutateAddPiece(100)
	}
	for i := 0; i < numWalls; i++ {
		board.mutateAddWall(100)
	}
	return board
}

func NewBoardFromString(desc string) (*Board, error) {
	s := int(math.Sqrt(float64(len(desc))))
	if s*s != len(desc) {
		return nil, fmt.Errorf("NewBoardFromString only supports square boards")
	}
	rows := make([]string, s)
	for i := range rows {
		rows[i] = desc[i*s : i*s+s]
	}
	return NewBoard(rows)
}

func NewBoard(desc []string) (*Board, error) {
	// determine board size
	h := len(desc)
	if h < MinBoardSize {
		return nil, fmt.Errorf("board height must be >= %d", MinBoardSize)
	}
	w := len(desc[0])
	if w < MinBoardSize {
		return nil, fmt.Errorf("board width must be >= %d", MinBoardSize)
	}

	// identify occupied cells and their labels
	occupied := make([]bool, w*h)
	positions := make(map[string][]int)
	var walls []int
	for y, row := range desc {
		for x, value := range row {
			label := string(value)
			if label == "." || label == "o" {
				continue
			}
			i := y*w + x
			occupied[i] = true
			if label == "x" {
				walls = append(walls, i)
			} else {
				positions[label] = append(positions[label], i)
			}

		}
	}

	// find and sort distinct piece labels
	labels := make([]string, 0, len(positions))
	for label := range positions {
		labels = append(labels, label)
	}
	sort.Strings(labels)

	// validate and create pieces
	pieces := make([]Piece, 0, len(labels))
	for _, label := range labels {
		ps := positions[label]
		if len(ps) < MinPieceSize {
			return nil, fmt.Errorf("piece %s length must be >= %d", label, MinPieceSize)
		}
		stride := ps[1] - ps[0]
		if stride != 1 && stride != w {
			return nil, fmt.Errorf("piece %s has invalid shape", label)
		}
		for i := 2; i < len(ps); i++ {
			if ps[i]-ps[i-1] != stride {
				return nil, fmt.Errorf("piece %s has invalid shape", label)
			}
		}
		dir := Horizontal
		if stride != 1 {
			dir = Vertical
		}
		pieces = append(pieces, Piece{ps[0], len(ps), dir})
	}

	// create board
	board := &Board{w, h, pieces, walls, occupied, MakeMemoKey(pieces)}
	return board, board.Validate()
}

func (board *Board) String() string {
	w := board.Width
	h := board.Height
	grid := make([]string, w*h)
	for i := range grid {
		grid[i] = "."
	}
	for _, i := range board.Walls {
		grid[i] = "x"
	}
	for i, piece := range board.Pieces {
		label := string('A' + i)
		idx := piece.Position
		stride := piece.Stride(w)
		for j := 0; j < piece.Size; j++ {
			grid[idx] = label
			idx += stride
		}
	}
	rows := make([]string, h)
	for y := 0; y < h; y++ {
		i := y * w
		rows[y] = strings.Join(grid[i:i+w], "")
	}
	return strings.Join(rows, "\n")
}

func (board *Board) Hash() string {
	w := board.Width
	h := board.Height
	grid := make([]rune, w*h)
	for i := range grid {
		grid[i] = '.'
	}
	for _, i := range board.Walls {
		grid[i] = 'x'
	}
	for i, piece := range board.Pieces {
		label := rune('A' + i)
		idx := piece.Position
		stride := 1
		if piece.Orientation == Vertical {
			stride = w
		}
		for j := 0; j < piece.Size; j++ {
			grid[idx] = label
			idx += stride
		}
	}
	return string(grid)
}

func (board *Board) Copy() *Board {
	w := board.Width
	h := board.Height
	pieces := make([]Piece, len(board.Pieces))
	walls := make([]int, len(board.Walls))
	occupied := make([]bool, len(board.occupied))
	memoKey := board.memoKey
	copy(pieces, board.Pieces)
	copy(walls, board.Walls)
	copy(occupied, board.occupied)
	return &Board{w, h, pieces, walls, occupied, memoKey}
}

func (board *Board) SortPieces() {
	a := board.Pieces[1:]
	sort.Slice(a, func(i, j int) bool {
		return a[i].Position < a[j].Position
	})
	board.memoKey = MakeMemoKey(board.Pieces)
}

func (board *Board) HasFullRowOrCol() bool {
	w := board.Width
	h := board.Height
	for y := 0; y < h; y++ {
		var size int
		for _, piece := range board.Pieces {
			if piece.Orientation == Horizontal && piece.Row(w) == y {
				size += piece.Size
			}

		}
		if size == w {
			return true
		}
	}
	for x := 0; x < w; x++ {
		var size int
		for _, piece := range board.Pieces {
			if piece.Orientation == Vertical && piece.Col(w) == x {
				size += piece.Size
			}
		}
		if size == h {
			return true
		}
	}
	return false
}

func (board *Board) Validate() error {
	w := board.Width
	h := board.Height
	pieces := board.Pieces

	// board size must be >= MinBoardSize
	if w < MinBoardSize {
		return fmt.Errorf("board width must be >= %d", MinBoardSize)
	}
	if h < MinBoardSize {
		return fmt.Errorf("board height must be >= %d", MinBoardSize)
	}

	// board must have at least one piece
	if len(pieces) < 1 {
		return fmt.Errorf("board must have at least one piece")
	}

	// board must have <= MaxPieces
	if len(pieces) > MaxPieces {
		return fmt.Errorf("board must have <= %d pieces", MaxPieces)
	}

	// primary piece must be horizontal
	if pieces[0].Orientation != Horizontal {
		return fmt.Errorf("primary piece must be horizontal")
	}

	// validate walls
	occupied := make([]bool, w*h)
	for _, i := range board.Walls {
		// wall must be inside the grid
		if i < 0 || i >= w*h {
			return fmt.Errorf("a wall is outside of the grid")
		}

		// walls must not intersect
		if occupied[i] {
			return fmt.Errorf("a wall intersects another wall")
		}
		occupied[i] = true
	}

	// validate pieces
	primaryRow := pieces[0].Row(w)
	for i, piece := range pieces {
		label := string('A' + i)
		row := piece.Row(w)
		col := piece.Col(w)

		// piece size must be >= MinPieceSize
		if piece.Size < MinPieceSize {
			return fmt.Errorf("piece %s must have size >= %d", label, MinPieceSize)
		}

		// no horizontal pieces can be on the same row as the primary piece
		if i > 0 && piece.Orientation == Horizontal && row == primaryRow {
			return fmt.Errorf("no horizontal pieces can be on the primary row")
		}

		// pieces must be contained within the grid
		if piece.Orientation == Horizontal {
			if row < 0 || row >= h || col < 0 || col+piece.Size > w {
				return fmt.Errorf("piece %s is outside of the grid", label)
			}
		} else {
			if col < 0 || col >= w || row < 0 || row+piece.Size > h {
				return fmt.Errorf("piece %s is outside of the grid", label)
			}
		}

		// pieces must not intersect
		idx := piece.Position
		stride := piece.Stride(w)
		for j := 0; j < piece.Size; j++ {
			if occupied[idx] {
				return fmt.Errorf("piece %s intersects with another piece", label)
			}
			occupied[idx] = true
			idx += stride
		}
	}

	return nil
}

func (board *Board) isOccupied(piece Piece) bool {
	idx := piece.Position
	stride := piece.Stride(board.Width)
	for i := 0; i < piece.Size; i++ {
		if board.occupied[idx] {
			return true
		}
		idx += stride
	}
	return false
}

func (board *Board) setOccupied(piece Piece, value bool) {
	idx := piece.Position
	stride := piece.Stride(board.Width)
	for i := 0; i < piece.Size; i++ {
		board.occupied[idx] = value
		idx += stride
	}
}

func (board *Board) addPiece(piece Piece) {
	i := len(board.Pieces)
	board.Pieces = append(board.Pieces, piece)
	board.setOccupied(piece, true)
	board.memoKey[i] = piece.Position
}

func (board *Board) AddPiece(piece Piece) bool {
	if board.isOccupied(piece) {
		return false
	}
	board.addPiece(piece)
	return true
}

func (board *Board) AddWall(i int) bool {
	if board.occupied[i] {
		return false
	}
	board.Walls = append(board.Walls, i)
	board.occupied[i] = true
	return true
}

func (board *Board) RemovePiece(i int) {
	board.setOccupied(board.Pieces[i], false)
	j := len(board.Pieces) - 1
	board.Pieces[i] = board.Pieces[j]
	board.memoKey[i] = board.Pieces[i].Position
	board.Pieces = board.Pieces[:j]
	board.memoKey[j] = 0
}

func (board *Board) RemoveLastPiece() {
	board.RemovePiece(len(board.Pieces) - 1)
}

func (board *Board) RemoveWall(i int) {
	board.occupied[board.Walls[i]] = false
	a := board.Walls
	a[i] = a[len(a)-1]
	a = a[:len(a)-1]
	board.Walls = a
}

func (board *Board) Target() int {
	w := board.Width
	piece := board.Pieces[0]
	row := piece.Row(w)
	return (row+1)*w - piece.Size
}

func (board *Board) Moves(buf []Move) []Move {
	moves := buf[:0]
	w := board.Width
	h := board.Height
	for i, piece := range board.Pieces {
		var stride, reverseSteps, forwardSteps int
		if piece.Orientation == Vertical {
			y := piece.Position / w
			reverseSteps = -y
			forwardSteps = h - piece.Size - y
			stride = w
		} else {
			x := piece.Position % w
			reverseSteps = -x
			forwardSteps = w - piece.Size - x
			stride = 1
		}
		// reverse (negative steps)
		idx := piece.Position - stride
		for steps := -1; steps >= reverseSteps; steps-- {
			if board.occupied[idx] {
				break
			}
			moves = append(moves, Move{i, steps})
			idx -= stride
		}
		// forward (positive steps)
		idx = piece.Position + piece.Size*stride
		for steps := 1; steps <= forwardSteps; steps++ {
			if board.occupied[idx] {
				break
			}
			moves = append(moves, Move{i, steps})
			idx += stride
		}
	}
	return moves
}

func (board *Board) DoMove(move Move) {
	piece := &board.Pieces[move.Piece]
	stride := piece.Stride(board.Width)

	idx := piece.Position
	for i := 0; i < piece.Size; i++ {
		board.occupied[idx] = false
		idx += stride
	}

	piece.Position += stride * move.Steps
	board.memoKey[move.Piece] = piece.Position

	idx = piece.Position
	for i := 0; i < piece.Size; i++ {
		board.occupied[idx] = true
		idx += stride
	}
}

func (board *Board) UndoMove(move Move) {
	board.DoMove(Move{move.Piece, -move.Steps})
}

func (board *Board) StateIterator() <-chan *Board {
	ch := make(chan *Board, 16)
	board = board.Copy()
	memo := NewMemo()
	var f func(int, int)
	f = func(depth, previousPiece int) {
		if !memo.Add(board.MemoKey(), 0) {
			return
		}
		ch <- board.Copy()
		for _, move := range board.Moves(nil) {
			if move.Piece == previousPiece {
				continue
			}
			board.DoMove(move)
			f(depth+1, move.Piece)
			board.UndoMove(move)
		}
		if depth == 0 {
			close(ch)
		}
	}
	go f(0, -1)
	return ch
}

func (board *Board) ReachableStates() int {
	var count int
	memo := NewMemo()
	var f func(int)
	f = func(previousPiece int) {
		if !memo.Add(board.MemoKey(), 0) {
			return
		}
		count++
		for _, move := range board.Moves(nil) {
			if move.Piece == previousPiece {
				continue
			}
			board.DoMove(move)
			f(move.Piece)
			board.UndoMove(move)
		}
	}
	f(-1)
	return count
}

func (board *Board) MemoKey() *MemoKey {
	return &board.memoKey
}

func (board *Board) Solve() Solution {
	return NewSolver(board).Solve()
}

func (board *Board) Unsolve() (*Board, Solution) {
	return NewUnsolver(board).Unsolve()
}

func (board *Board) UnsafeSolve() Solution {
	return NewSolver(board).UnsafeSolve()
}

func (board *Board) UnsafeUnsolve() (*Board, Solution) {
	return NewUnsolver(board).UnsafeUnsolve()
}

func (board *Board) Render() image.Image {
	return renderBoard(board)
}

func (board *Board) Impossible() bool {
	return theStaticAnalyzer.Impossible(board)
}

func (board *Board) BlockedSquares() []int {
	return theStaticAnalyzer.BlockedSquares(board)
}

func (board *Board) Canonicalize() *Board {
	bestKey := board.memoKey
	bestBoard := board.Copy()
	for b := range board.StateIterator() {
		if b.memoKey.Less(&bestKey, true) {
			bestKey = b.memoKey
			bestBoard = b.Copy()
		}
	}
	bestBoard.SortPieces()
	return bestBoard
}

// random board mutation below

func (board *Board) Energy() float64 {
	solution := board.Solve()
	if !solution.Solvable {
		return 1
	}
	e := float64(solution.NumMoves)
	e += float64(solution.NumSteps) / 100
	return -e
}

type UndoFunc func()

func (board *Board) Mutate() UndoFunc {
	const maxAttempts = 100
	for {
		var undo UndoFunc
		switch rand.Intn(7 + 3) {
		case 0:
			undo = board.mutateAddPiece(maxAttempts)
		case 1:
			undo = board.mutateAddWall(maxAttempts)
		case 2:
			undo = board.mutateRemovePiece()
		case 3:
			undo = board.mutateRemoveWall()
		case 4:
			undo = board.mutateRemoveAndAddPiece(maxAttempts)
		case 5:
			undo = board.mutateRemoveAndAddWall(maxAttempts)
		default:
			undo = board.mutateMakeMove()
		}
		if undo != nil {
			return undo
		}
	}
}

func (board *Board) mutateMakeMove() UndoFunc {
	moves := board.Moves(nil)
	if len(moves) == 0 {
		return nil
	}
	move := moves[rand.Intn(len(moves))]
	board.DoMove(move)
	return func() {
		board.UndoMove(move)
	}
}

func (board *Board) mutateAddPiece(maxAttempts int) UndoFunc {
	if len(board.Pieces) >= 8 {
		return nil
	}
	piece, ok := board.randomPiece(maxAttempts)
	if !ok {
		return nil
	}
	i := len(board.Pieces)
	board.AddPiece(piece)
	return func() {
		board.RemovePiece(i)
	}
}

func (board *Board) mutateAddWall(maxAttempts int) UndoFunc {
	if len(board.Walls) >= 0 {
		return nil
	}
	wall, ok := board.randomWall(maxAttempts)
	if !ok {
		return nil
	}
	i := len(board.Walls)
	board.AddWall(wall)
	return func() {
		board.RemoveWall(i)
	}
}

func (board *Board) mutateRemovePiece() UndoFunc {
	// never remove the primary piece
	if len(board.Pieces) < 2 {
		return nil
	}
	i := rand.Intn(len(board.Pieces)-1) + 1
	piece := board.Pieces[i]
	board.RemovePiece(i)
	return func() {
		board.AddPiece(piece)
	}
}

func (board *Board) mutateRemoveWall() UndoFunc {
	if len(board.Walls) == 0 {
		return nil
	}
	i := rand.Intn(len(board.Walls))
	wall := board.Walls[i]
	board.RemoveWall(i)
	return func() {
		board.AddWall(wall)
	}
}

func (board *Board) mutateRemoveAndAddPiece(maxAttempts int) UndoFunc {
	undoRemove := board.mutateRemovePiece()
	if undoRemove == nil {
		return nil
	}
	undoAdd := board.mutateAddPiece(maxAttempts)
	if undoAdd == nil {
		return undoRemove
	}
	return func() {
		undoAdd()
		undoRemove()
	}
}

func (board *Board) mutateRemoveAndAddWall(maxAttempts int) UndoFunc {
	undoRemove := board.mutateRemoveWall()
	if undoRemove == nil {
		return nil
	}
	undoAdd := board.mutateAddWall(maxAttempts)
	if undoAdd == nil {
		return undoRemove
	}
	return func() {
		undoAdd()
		undoRemove()
	}
}

func (board *Board) randomPiece(maxAttempts int) (Piece, bool) {
	w := board.Width
	h := board.Height
	for i := 0; i < maxAttempts; i++ {
		size := 2 + rand.Intn(2) // TODO: weighted?
		orientation := Orientation(rand.Intn(2))
		var x, y int
		if orientation == Vertical {
			x = rand.Intn(w)
			y = rand.Intn(h - size + 1)
		} else {
			x = rand.Intn(w - size + 1)
			y = rand.Intn(h)
		}
		position := y*w + x
		piece := Piece{position, size, orientation}
		if !board.isOccupied(piece) {
			return piece, true
		}
	}
	return Piece{}, false
}

func (board *Board) randomWall(maxAttempts int) (int, bool) {
	n := board.Width * board.Height
	for i := 0; i < maxAttempts; i++ {
		p := rand.Intn(n)
		if !board.occupied[p] {
			return p, true
		}
	}
	return 0, false
}
