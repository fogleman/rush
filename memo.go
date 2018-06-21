package rush

const MaxPieces = 32

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
}

func NewMemo() *Memo {
	data := make(map[MemoKey]int)
	return &Memo{data}
}

func (memo *Memo) Size() int {
	return len(memo.data)
}

func (memo *Memo) Add(key *MemoKey, depth int) bool {
	if before, ok := memo.data[*key]; ok && before >= depth {
		return false
	}
	memo.data[*key] = depth
	return true
}
