package store

type MemoryStore struct {
	chunkSize int64
	chunks    map[string][]byte
}

func (s *MemoryStore) Put(chunkID string, b []byte) (int, error) {
	// fmt.Printf(">>> PUT(%s) [%x]\n", chunkID, b)
	s.chunks[chunkID] = b
	return len(b), nil
}

func (s *MemoryStore) Get(chunkID string) ([]byte, error) {
	if ch, ok := s.chunks[chunkID]; ok {
		// fmt.Printf(">>> GET(%s) [%x]\n", chunkID, ch)
		return ch, nil
	}
	return make([]byte, s.chunkSize), nil
}
