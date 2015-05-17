package base

func makeBufferUnit(sz int) []byte {
	buf := make([]byte, sz)

	return buf
}

func makeBufferPool(n int, sz int) [][]byte {
	pool := make([][]byte, n)

	for i := 0; i < n; i++ {
		pool[i] = makeBufferUnit(sz)
	}

	return pool
}
