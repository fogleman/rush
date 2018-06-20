package rush

const MaxPieces = 32

type MemoKey [MaxPieces]int

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

func (memo *Memo) Add(key MemoKey, depth int) bool {
	if before, ok := memo.data[key]; ok && before >= depth {
		return false
	}
	memo.data[key] = depth
	return true
}
