package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/GitbookIO/syncgroup"
	"github.com/libopenstorage/openstorage/volume/drivers/buse"
	"github.com/sirupsen/logrus"

	"github.com/geoah/go-block-gdrive/chunks"
	"github.com/geoah/go-block-gdrive/store"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	logrus.Infof("Initializing nbd...")
	if _, err := os.Stat("/usr/sbin/modprobe"); err == nil {
		exec.Command("/usr/sbin/modprobe", "nbd").Output()
	} else {
		exec.Command("/sbin/modprobe", "nbd").Output()
	}

	for i := 0; ; i++ {
		dev := fmt.Sprintf("/dev/nbd%d", i)
		if _, err := os.Stat(dev); os.IsNotExist(err) {
			break
		}

		syscall.Unmount(dev, syscall.MNT_DETACH)

		logrus.Debugf("Cleaning up %s", dev)

		if f, err := os.Open(dev); err == nil {
			ioctl(f.Fd(), NBD_DISCONNECT, 0)
			ioctl(f.Fd(), NBD_CLEAR_QUE, 0)
			ioctl(f.Fd(), NBD_CLEAR_SOCK, 0)

			f.Close()
		}
	}

	ms := &store.MemoryStore{
		chunkSize: 4096,
		chunks:    map[string][]byte{},
	}
	ld := &chunks.ChunkedDevice{
		store:     ms,
		chunkSize: 4096,
		lock:      syncgroup.NewMutexGroup(),
	}

	logrus.Infof("Creating device...")
	nbd := buse.Create(ld, "local-aaa", 4096*1000) // TODO Fix hardcoded size

	logrus.Infof("Creating to device...")
	dev, err := nbd.Connect()
	if err != nil {
		logrus.WithError(err).Fatalf("Could not connect device")
	}

	logrus.Infof("Connected to %s", dev)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		logrus.Infof("Disconnecting device")
		nbd.Disconnect()
		logrus.Infof("Done cleaning up NBD devices")
		os.Exit(0)
	}()

	defer func() {
		if r := recover(); r != nil {
			logrus.WithField("r", r).Error("Recovered")
		}
	}()

	c := make(chan struct{})
	<-c
}
