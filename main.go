package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/geoah/go-block-gdrive/chunks"
	"github.com/geoah/go-block-gdrive/nbd"
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
			nbd.IOCTL(f.Fd(), nbd.NBD_DISCONNECT, 0)
			nbd.IOCTL(f.Fd(), nbd.NBD_CLEAR_QUE, 0)
			nbd.IOCTL(f.Fd(), nbd.NBD_CLEAR_SOCK, 0)

			f.Close()
		}
	}

	logrus.Infof("Creating device and interface...")
	// st, _ := store.NewMemoryStore(4096)
	st, _ := store.NewFileStore("/tmp")
	ld, _ := chunks.NewChunkedDevice(st, 4096*1000, 4096)
	ni := nbd.Create(ld, "local-aaa", 4096*1000)

	logrus.Infof("Creating to device...")
	dev, err := ni.Connect()
	if err != nil {
		logrus.WithError(err).Fatalf("Could not connect device")
	}

	logrus.Infof("Connected to %s", dev)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		<-c
		logrus.Infof("Disconnecting device")
		ni.Disconnect()
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

// IOCTL helper function
func IOCTL(a1, a2, a3 uintptr) (err error) {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, a1, a2, a3)
	if errno != 0 {
		err = errno
	}
	return err
}
