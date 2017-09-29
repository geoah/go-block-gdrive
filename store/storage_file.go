package store

import (
	"io/ioutil"
	"os"
	"path"
)

func NewFileStore(path string) (Store, error) {
	s := &FileStore{
		path: path,
	}
	return s, nil
}

type FileStore struct {
	path string
}

func (s *FileStore) Put(chunkID string, b []byte) (int, error) {
	fn := s.getFilename(chunkID)
	if err := ioutil.WriteFile(fn, b, 0644); err != nil {
		return 0, err
	}

	return len(b), nil
}

func (s *FileStore) Get(chunkID string) ([]byte, error) {
	fn := s.getFilename(chunkID)
	fl, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer fl.Close()

	data, err := ioutil.ReadAll(fl)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *FileStore) getFilename(chunkID string) string {
	return path.Join(s.path, chunkID)
}
