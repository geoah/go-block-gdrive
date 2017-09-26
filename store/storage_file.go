package store

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

type FileStore struct {
	path string
}

func (s *FileStore) Put(chunkID string, b []byte) (int, error) {
	fn := s.getFilename(chunkID)
	fl, err := os.Create(fn)
	if err != nil {
		return 0, err
	}
	defer fl.Close()

	n, err := fl.Write(b)
	if err != nil {
		logrus.Warningf("error writing file %s, err: %s", chunkID, err)
		return 0, err
	}

	return n, nil
}

func (s *FileStore) Get(chunkID string) ([]byte, error) {
	fn := s.getFilename(chunkID)
	fl, err := os.Open(fn)
	if err != nil {
		return nil, nil
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
