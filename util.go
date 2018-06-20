package rush

func updateOccupied(occupied []bool, stride, start, size int, value bool) {
	idx := start
	for i := 0; i < size; i++ {
		occupied[idx] = value
		idx += stride
	}
}
