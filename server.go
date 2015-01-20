package main

import (
	"encoding/binary"
	"fmt"
	ss "github.com/shadowsocks/shadowsocks-go/shadowsocks"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
	"syscall"
)

//type Clients []Clienter
//func ReloadServer(cs *Clients) {
//	var k int64
//	// add new user to listen...
//	for _, c := range cs {
//		passwdManager.updatePortPasswd(c)
//	}
//	// close deleted/deactive user's port
//	for _, k = range passwdManager.getkeys() {
//		user := User{}
//		u, err := user.GetUserById(k)
//		if (err != nil && err == ErrUserNotExist) || !u.IsActive {
//			passwdManager.del(k)
//		}
//	}
//}

const logCntDelta = 100

var (
	connCnt        int
	nextLogConnCnt int = logCntDelta
	passwdManager      = PasswdManager{portListener: map[int64]*PortListener{}}
)

type Clienter interface {
	GetUsername() string
	GetId() int64
	GetPassword() string
	GetMethod() string
	GetPort() int
	UpdateInBytes(string, int64) bool
	UpdateOutBytes(string, int64) bool
	FirstVisitToday() bool
	GetLimited() bool
	GetUsedBytes() (int64, int64)
}

type PortListener struct {
	client   Clienter
	listener net.Listener
}

type PasswdManager struct {
	sync.Mutex
	portListener map[int64]*PortListener
}

func (pm *PasswdManager) add(client Clienter, listener net.Listener) {
	pm.Lock()
	pm.portListener[client.GetId()] = &PortListener{client, listener}
	pm.Unlock()
}

func (pm *PasswdManager) get(id int64) (pl *PortListener, ok bool) {
	pm.Lock()
	pl, ok = pm.portListener[id]
	pm.Unlock()
	return
}

func (pm *PasswdManager) getkeys() []int64 {
	pm.Lock()
	keys := make([]int64, 0, len(pm.portListener))
	for k := range pm.portListener {
		keys = append(keys, k)
	}
	pm.Unlock()
	return keys
}

func (pm *PasswdManager) del(id int64) {
	pl, ok := pm.get(id)
	if !ok {
		return
	}
	pl.listener.Close()
	pm.Lock()
	delete(pm.portListener, id)
	pm.Unlock()
}

// Update port password would first close a port and restart listening on that
// port. A different approach would be directly change the password used by
// that port, but that requires **sharing** password between the port listener
// and password manager.
func (pm *PasswdManager) updatePortPasswd(client Clienter) {
	if pl, ok := pm.get(client.GetId()); !ok {
		log.Printf("new port %v added\n", client.GetPort())
	} else {
		if pl.client.GetPassword() == client.GetPassword() &&
			pl.client.GetMethod() == client.GetMethod() &&
			pl.client.GetPort() == client.GetPort() {
			return
		}
		log.Printf("closing port %s to update password\n", client.GetPort())
		pl.listener.Close()
	}
	// run will add the new port listener to passwdManager.
	// So there maybe concurrent access to passwdManager and we need lock to protect it.
	go run(client)
}

func run(client Clienter) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(client.GetPort()))
	if err != nil {
		log.Printf("error listening port %v: %v\n", client.GetPort(), err)
		return
	}
	passwdManager.add(client, ln)
	var cipher *ss.Cipher
	log.Printf("server listening port %v ...\n", client.GetPort())
	for {
		conn, err := ln.Accept()
		if err != nil {
			// listener maybe closed to update password
			log.Printf("accept error: %v\n", err)
			return
		}
		// Creating cipher upon first connection.
		if cipher == nil {
			log.Println("creating cipher for port:", client.GetPort())
			cipher, err = ss.NewCipher(client.GetMethod(), client.GetPassword())
			if err != nil {
				log.Printf("Error generating cipher for port: %s %v\n", client.GetPort(), err)
				conn.Close()
				continue
			}
		}
		// 用户本月流量已经用完if user get count bytes, then writeLimitPage
		clientConn := ss.NewConn(conn, cipher.Copy())
		if client.GetLimited() {
			WriteErrorPage(clientConn, client.GetUsername())
			clientConn.Close()
			conn.Close()
			continue
		}
		go handleConnection(client, clientConn)
	}
}

func handleConnection(client Clienter, conn *ss.Conn) {
	var host string

	connCnt++ // this maybe not accurate, but should be enough
	if connCnt-nextLogConnCnt >= 0 {
		// XXX There's no xadd in the atomic package, so it's difficult to log
		// the message only once with low cost. Also note nextLogConnCnt maybe
		// added twice for current peak connection number level.
		log.Printf("Number of client connections reaches %d\n", nextLogConnCnt)
		nextLogConnCnt += logCntDelta
	}

	// function arguments are always evaluated, so surround debug statement
	// with if statement
	if app.config.debug {
		log.Printf("new client %s->%s\n", conn.RemoteAddr().String(), conn.LocalAddr())
	}
	closed := false
	defer func() {
		if app.config.debug {
			log.Printf("closed pipe %s<->%s\n", conn.RemoteAddr(), host)
		}
		connCnt--
		if !closed {
			conn.Close()
		}
	}()

	host, extra, err := getRequest(conn)
	if err != nil {
		log.Println("error getting request", conn.RemoteAddr(), conn.LocalAddr(), err)
		return
	}

	// 如果是每天第一次访问，则显示当前已经使用的流量
	if client.FirstVisitToday() {
		used, count := client.GetUsedBytes()
		WriteInfoPage(conn, used, count)
		client.UpdateOutBytes(app.config.http_host, 930)
		return
	}
	log.Println("connecting", host)
	remote, err := net.Dial("tcp", host)
	if err != nil {
		if ne, ok := err.(*net.OpError); ok && (ne.Err == syscall.EMFILE || ne.Err == syscall.ENFILE) {
			// log too many open file error
			// EMFILE is process reaches open file limits, ENFILE is system limit
			log.Println("dial error:", err)
		} else {
			log.Println("error connecting to:", host, err)
		}
		return
	}
	defer func() {
		if !closed {
			remote.Close()
		}
	}()
	// write extra bytes read from
	if extra != nil {
		// debug.Println("getRequest read extra data, writing to remote, len", len(extra))
		if _, err = remote.Write(extra); err != nil {
			log.Println("write request extra error:", err)
			return
		}
	}
	if app.config.debug {
		log.Printf("piping %s<->%s", conn.RemoteAddr(), host)
	}

	go PipeThenClose(conn, remote, ss.SET_TIMEOUT, client, host)
	PipeThenClose(remote, conn, ss.NO_TIMEOUT, client, host)
	closed = true
	return
}

func getRequest(conn *ss.Conn) (host string, extra []byte, err error) {
	const (
		idType  = 0 // address type index
		idIP0   = 1 // ip addres start index
		idDmLen = 1 // domain address length index
		idDm0   = 2 // domain address start index

		typeIPv4 = 1 // type is ipv4 address
		typeDm   = 3 // type is domain address
		typeIPv6 = 4 // type is ipv6 address

		lenIPv4   = 1 + net.IPv4len + 2 // 1addrType + ipv4 + 2port
		lenIPv6   = 1 + net.IPv6len + 2 // 1addrType + ipv6 + 2port
		lenDmBase = 1 + 1 + 2           // 1addrType + 1addrLen + 2port, plus addrLen
	)

	// buf size should at least have the same size with the largest possible
	// request size (when addrType is 3, domain name has at most 256 bytes)
	// 1(addrType) + 1(lenByte) + 256(max length address) + 2(port)
	buf := make([]byte, 260)
	var n int
	// read till we get possible domain length field
	ss.SetReadTimeout(conn)
	if n, err = io.ReadAtLeast(conn, buf, idDmLen+1); err != nil {
		return
	}

	reqLen := -1
	switch buf[idType] {
	case typeIPv4:
		reqLen = lenIPv4
	case typeIPv6:
		reqLen = lenIPv6
	case typeDm:
		reqLen = int(buf[idDmLen]) + lenDmBase
	default:
		err = fmt.Errorf("addr type %d not supported", buf[idType])
		return
	}

	if n < reqLen { // rare case
		ss.SetReadTimeout(conn)
		if _, err = io.ReadFull(conn, buf[n:reqLen]); err != nil {
			return
		}
	} else if n > reqLen {
		// it's possible to read more than just the request head
		extra = buf[reqLen:n]
	}

	// Return string for typeIP is not most efficient, but browsers (Chrome,
	// Safari, Firefox) all seems using typeDm exclusively. So this is not a
	// big problem.
	switch buf[idType] {
	case typeIPv4:
		host = net.IP(buf[idIP0 : idIP0+net.IPv4len]).String()
	case typeIPv6:
		host = net.IP(buf[idIP0 : idIP0+net.IPv6len]).String()
	case typeDm:
		host = string(buf[idDm0 : idDm0+buf[idDmLen]])
	}
	// parse port
	port := binary.BigEndian.Uint16(buf[reqLen-2 : reqLen])
	host = net.JoinHostPort(host, strconv.Itoa(int(port)))
	return
}
