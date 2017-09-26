package store

type Store interface {
	Put(chunkID string, b []byte) (int, error)
	Get(chunkID string) ([]byte, error)
}
