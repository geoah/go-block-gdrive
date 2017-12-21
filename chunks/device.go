package chunks

import (
	"fmt"
	"math"
	"sync"

	"github.com/GitbookIO/syncgroup"
	"github.com/libopenstorage/openstorage/volume/drivers/buse"
	"github.com/sirupsen/logrus"

	"github.com/geoah/go-block-gdrive/lru"
	"github.com/geoah/go-block-gdrive/store"
)

const (
	defaultDirtyThreshold = 0.20
)

func NewChunkedDevice(store store.Store, chunkSize, nChunks int64, cacheSize int, dirtyThreshold float64) (buse.Device, error) {
	if dirtyThreshold > 1 || dirtyThreshold <= 0 {
		dirtyThreshold = defaultDirtyThreshold
	}
	dirtySize := math.Floor(dirtyThreshold * float64(nChunks))

	d := &ChunkedDevice{
		nChunks:     nChunks,
		chunkSize:   chunkSize,
		store:       store,
		lock:        syncgroup.NewMutexGroup(),
		lru:         lru.NewLRU(cacheSize),
		dirtyChunks: lru.NewLRU(dirtySize),
	}
	return d, nil
}

type ChunkedDevice struct {
	nChunks     int64
	chunkSize   int64
	store       store.Store
	lock        *syncgroup.MutexGroup
	cache       *lru.LRU
	dirtyChunks *lru.LRU
}

type chunk struct {
	id   int64
	data []byte
}

// Push a chunk to the store and optionaly remove it from the
// dirty cache.
func (d *ChunkedDevice) pushChunkToStore(c *chunk, removeDirty bool) {
	d.lock.RLock(c.id)
	defer d.lock.RUnlock(c.id)

	d.store.Put(c.id, c.data)
	if removeDirty {
		d.dirtyChunks.Remove(c.id)
	}
}

// Look up the chunk in the cache. If found return it.
// If not create a new chunk by getting the data from
// the store. Put in the cache and push any cache left overs
// back.
func (d *ChunkedDevice) fetchChunk(id int64) *chunk {
	d.lock.Lock(id)
	defer d.lock.Unlock(id)

	c, ok := d.cache.Get(id)
	if !ok {
		c = &chunk{
			id:   id,
			data: d.store.Get(id),
		}
		if v, ok := d.cache.Put(id, c); ok {
			d.pushChunkToStore(v, true)
		}
	}
	return c.(*chunk)
}

// Cleanup all dirty chunks by pushing them back
// to store and then flush the dirty cache.
func (d *ChunkedDevice) cleanup() (n int) {
	d.dirtyChunks.Flush(func(c interface{}) {
		d.pushChunkToStore(c.(*chunk), false)
		n++
	})

	return
}

// Mark a chunk as dirty. If we are over our capacity run a cleanup.
// Return true if we had to run a cleanup.
func (d *ChunkedDevice) markAsDirty(c *chunk) bool {
	v, ok := d.dirtyChunks.Put(c.id, c)
	if ok {
		d.pushChunkToStore(c, false)
		d.cleanup()
	}
	return ok
}

func (d *ChunkedDevice) basicIO(b []byte, off int64, write bool) (int, error) {
	n := int64(len(b))
	first := off / d.chunkSize
	nblocks := n / d.chunkSize
	if nblocks > d.nChunks {
		return -1, fmt.Errorf("Requested size exceeds device size")
	}

	cId := first
	bS := 0
	for n > 0 {
		c := d.fetchChunk(cId)
		left := d.chunkSize
		if n < d.chunkSize {
			left = n
		}
		if cId != first {
			off = 0
		}
		r := 0
		if write {
			r = copy(c.data[off:left], b[bS:])
			d.markAsDirty(c)
		} else {
			r = copy(b[bS:], c.data[off:left])
		}
		bS += r
		n -= r
		cId++
	}

	return n, nil
}

func (d *ChunkedDevice) ReadAt(b []byte, off int64) (int, error) {
	return d.basicIO(b, off, false)
}

func (d *ChunkedDevice) WriteAt(b []byte, off int64) (int, error) {
	return d.basicIO(b, off, true)
}
