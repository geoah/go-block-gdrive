package chunks

import (
	"fmt"

	"github.com/GitbookIO/syncgroup"
	"github.com/libopenstorage/openstorage/volume/drivers/buse"
	"github.com/sirupsen/logrus"

	"github.com/geoah/go-block-gdrive/store"
)

func NewChunkedDevice(store store.Store, size, chunkSize int64) (buse.Device, error) {
	d := &ChunkedDevice{
		size:      size,
		chunkSize: chunkSize,
		store:     store,
		lock:      syncgroup.NewMutexGroup(),
	}
	return d, nil
}

type ChunkedDevice struct {
	size      int64
	chunkSize int64
	store     store.Store
	lock      *syncgroup.MutexGroup
}

func (d *ChunkedDevice) ReadAt(b []byte, off int64) (int, error) {
	logrus.WithField("offset", off).WithField("size", len(b)).Debugf("ReadAt")
	cr, _ := d.getChunkInfo(off, int64(len(b)))
	d.lock.RLock(cr.chunkID)
	cb, err := d.store.Get(cr.chunkID)
	if err != nil {
		logrus.WithError(err).Warningf("error reading chunk %s", cr.chunkID)
		return 0, err
	}
	if cb != nil {
		// TODO Check if this is over the size of the device?
		// fmt.Printf("___ [%d:%d] vs [%d:%d]\n", off, off+int64(len(b)), cr.offsetStart, cr.chunkOffsetEnd)
		// fmt.Printf(">>> GET(%s) [%x]\n", cr.chunkID, cb[cr.chunkOffset:cr.chunkOffsetEnd])
		n := copy(b, cb[cr.chunkOffset:cr.chunkOffsetEnd])
	}
	d.lock.RUnlock(cr.chunkID)
	return n, nil
}

func (d *ChunkedDevice) WriteAt(b []byte, off int64) (int, error) {
	logrus.WithField("offset", off).WithField("size", len(b)).Debugf("WriteAt")
	size := int64(len(b))
	left := size
	coff := off
	boff := int64(0) // index up to which we have written
	n := 0
	// find first chunk we need to write to
	for {
		// update how much is left to write
		csize := left
		if left > d.chunkSize {
			csize = d.chunkSize
			left = left - d.chunkSize
		}
		// find the chunk we need to write to
		cr, _ := d.getChunkInfo(coff, csize)
		// update the offset for the next write
		coff += csize
		// logrus.
		// 	WithField("chunkID", cr.chunkID).
		// 	WithField("csize", csize).
		// 	WithField("coff", coff).
		// 	WithField("left", left).
		// 	Debugf("> Writing to chunk")
		// lock and write chunk
		d.lock.Lock(cr.chunkID)
		nn, err := d.store.Put(cr.chunkID, b[boff:boff+csize])
		if err != nil {
			logrus.WithError(err).Warningf("error writing chunk %s", cr.chunkID)
			return 0, err
		}
		n += nn
		// unlock
		d.lock.Unlock(cr.chunkID)
		// update boff
		boff += csize
		// update left
		left -= csize
		// are we dont yet?
		if left == 0 {
			break
		}
		// TODO remove - just in case check
		if left < 0 {
			panic("left < 0")
		}
	}

	return n, nil
}

func (d *ChunkedDevice) getChunkInfo(off int64, size int64) (Info, error) {
	chunkOff := int64(0)
	if off > 0 {
		chunkOff = off % d.chunkSize
	}

	part := int64(0)
	if off > 0 {
		part = int64(off / d.chunkSize)
	}

	offStart := off - chunkOff
	offEnd := offStart + d.chunkSize
	ci := Info{
		offsetStart:    offStart,
		offsetEnd:      offEnd,
		chunkID:        fmt.Sprintf("part-%d", part),
		chunkOffset:    chunkOff,
		chunkOffsetEnd: chunkOff + int64(size),
	}

	// logrus.
	// 	WithField("off", off).
	// 	WithField("size", size).
	// 	WithField("offsetStart", ci.offsetStart).
	// 	WithField("offsetEnd", ci.offsetEnd).
	// 	WithField("chunkID", ci.chunkID).
	// 	WithField("chunkOffset", ci.chunkOffset).
	// 	WithField("chunkOffsetEnd", ci.chunkOffsetEnd).
	// 	Debug("getChunkInfo")

	return ci, nil
}
