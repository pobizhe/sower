package proxy

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/glog"
)

//go:generate stringer -type=netType $GOFILE
type netType int

const (
	QUIC netType = iota
	KCP
	TCP
)

func relay(conn1, conn2 net.Conn) {
	wg := &sync.WaitGroup{}
	exitFlag := new(int32)
	wg.Add(2)
	go redirect(conn1, conn2, wg, exitFlag)
	redirect(conn2, conn1, wg, exitFlag)
	wg.Wait()
}

func redirect(conn1, conn2 net.Conn, wg *sync.WaitGroup, exitFlag *int32) {
	if _, err := io.Copy(conn2, conn1); err != nil && (atomic.LoadInt32(exitFlag) == 0) {
		glog.V(1).Infof("%s<>%s -> %s<>%s: %s", conn1.RemoteAddr(), conn1.LocalAddr(), conn2.LocalAddr(), conn2.RemoteAddr(), err)
	}

	// wakeup all conn goroutine
	atomic.AddInt32(exitFlag, 1)
	now := time.Now()
	conn1.SetDeadline(now)
	conn2.SetDeadline(now)
	wg.Done()
}
