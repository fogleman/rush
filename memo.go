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

func (table *Memo) Add(key MemoKey, depth int) bool {
	if before, ok := table.data[key]; ok && before >= depth {
		return false
	}
	table.data[key] = depth
	return true
}
