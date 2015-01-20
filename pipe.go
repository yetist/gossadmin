package main

import (
	ss "github.com/shadowsocks/shadowsocks-go/shadowsocks"
	"net"
)

const bufSize = 4096
const nBuf = 2048

var pipeBuf = ss.NewLeakyBuf(nBuf, bufSize)

// PipeThenClose copies data from src to dst, closes dst when done.
func PipeThenClose(src, dst net.Conn, timeoutOpt int, client Clienter, host string) {
	defer dst.Close()
	buf := pipeBuf.Get()
	var l int64
	defer func() {
		pipeBuf.Put(buf)
		if timeoutOpt == ss.SET_TIMEOUT {
			client.UpdateInBytes(host, l)
		} else {
			client.UpdateOutBytes(host, l)
		}
	}()
	for {
		if timeoutOpt == ss.SET_TIMEOUT {
			ss.SetReadTimeout(src)
		}
		n, err := src.Read(buf)
		l += int64(n)
		// read may return EOF with n > 0
		// should always process n > 0 bytes before handling error
		if n > 0 {
			if _, err = dst.Write(buf[0:n]); err != nil {
				ss.Debug.Println("write:", err)
				break
			}
		}
		if err != nil {
			// Always "use of closed network connection", but no easy way to
			// identify this specific error. So just leave the error along for now.
			// More info here: https://code.google.com/p/go/issues/detail?id=4373
			/*
				if bool(Debug) && err != io.EOF {
					Debug.Println("read:", err)
				}
			*/
			break
		}
	}
}
