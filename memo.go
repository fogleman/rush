package rush

type MemoKey [MaxPieces]int

func MakeMemoKey(pieces []Piece) MemoKey {
	var key MemoKey
	for i, piece := range pieces {
		key[i] = piece.Position
	}
	return key
}

type Memo struct {
	data map[MemoKey]int
	hits uint64
}

func NewMemo() *Memo {
	data := make(map[MemoKey]int)
	return &Memo{data, 0}
}

func (memo *Memo) Size() int {
	return len(memo.data)
}

func (memo *Memo) Hits() uint64 {
	return memo.hits
}

func (memo *Memo) Add(key *MemoKey, depth int) bool {
	memo.hits++
	if before, ok := memo.data[*key]; ok && before >= depth {
		return false
	}
	memo.data[*key] = depth
	return true
}
