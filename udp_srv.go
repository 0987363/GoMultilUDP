// +build linux

package main

/*
#include <sys/types.h>
#include <sys/socket.h>
#include <stdlib.h>
*/
import "C"

import (
    "net"
    "os"
    "syscall"
	"fmt"
    "runtime"
	"time"
	"unsafe"
)

/*

Steps to get transparent proxying working on Linux:

1) Run these commands (as root):

iptables -t mangle -N DEMUX
iptables -t mangle -A DEMUX --jump MARK --set-mark 0x1
iptables -t mangle -A DEMUX --jump ACCEPT
ip rule add fwmark 0x1 lookup 100
ip route add local 0.0.0.0/0 dev lo table 100

2) Run the following commands - note that there's one for each interface/port
that you forward to:

iptables -t mangle -A OUTPUT --protocol tcp --out-interface eth0 --sport 22 --jump DEMUX
iptables -t mangle -A OUTPUT --protocol tcp --out-interface eth0 --sport 8080 --jump DEMUX

3) Finally, run demux (needs to be as root to use transparent proxying):

sudo ./demux -p 5555 --transparent=true \
--http-destination=<eth0 address>:8080 \
--ssh-destination=<eth0 address>:22

Note that the various destination addresses must be specified with the same
address as the interface you gave in step 2.

*/

func openTransparent(sip string, listen_port int) (net.Conn, error) {
    family := syscall.AF_INET
    sotype := syscall.SOCK_DGRAM
    proto := syscall.IPPROTO_UDP

	s, err := syscall.Socket(family, sotype, proto)
    if err != nil {
        syscall.CloseOnExec(s)
        return nil, err
    }

	err = syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
	if err != nil {
		fmt.Println("set multile port failed")
        return nil, err
    }
	fmt.Println("setsockoptint err:", err)

	var v = C.int(1)
	C.setsockopt(C.int(s), C.SOL_SOCKET, C.SO_REUSEPORT, unsafe.Pointer(&v), C.socklen_t(unsafe.Sizeof(v)));

    laddr, err := ipToSockaddr(family, sip, listen_port)
    if err != nil {
        return nil, err
    }

    if err := syscall.Bind(s, laddr); err != nil {
		fmt.Println("bind address failed")
        return nil, err
    }

    // Convert socket to *os.File
    // Note that net.FileConn() returns a copy of the socket, so we close this
    // File on return
    f := os.NewFile(uintptr(s), "dial")
    defer f.Close()
	defer syscall.Close(s)

    // Make a net.Conn from our file
    c, err := net.FileConn(f)
    if err != nil {
        fmt.Println("err: FileConn")
        return nil, err
    }

	fmt.Println("Created socket to ip:", sip, ":", listen_port)
    return c, nil
}

func openTransparent2(sip string, listen_port int) (*net.UDPConn, error) {
    family := syscall.AF_INET
    sotype := syscall.SOCK_DGRAM
    proto := syscall.IPPROTO_UDP

	s, err := syscall.Socket(family, sotype, proto)
    if err != nil {
        syscall.CloseOnExec(s)
        return nil, err
    }

	err = syscall.SetsockoptInt(s, syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1)
	if err != nil {
		fmt.Println("set multile port failed")
        return nil, err
    }
	fmt.Println("setsockoptint err:", err)

	var v = C.int(1)
	C.setsockopt(C.int(s), C.SOL_SOCKET, C.SO_REUSEPORT, unsafe.Pointer(&v), C.socklen_t(unsafe.Sizeof(v)));

    laddr, err := ipToSockaddr(family, sip, listen_port)
    if err != nil {
        return nil, err
    }

    if err := syscall.Bind(s, laddr); err != nil {
		fmt.Println("bind address failed")
        return nil, err
    }

    f := os.NewFile(uintptr(s), "dial")
    defer f.Close()
	defer syscall.Close(s)

	cc, err := net.FilePacketConn(f)
	if err != nil {
		return nil, err
	}
	return cc.(*net.UDPConn), nil
}

func sysSocket(family, sotype, proto int) (int, error) {
    syscall.ForkLock.RLock()
	s, err := syscall.Socket(family, sotype, proto)
    if err != nil {
        syscall.CloseOnExec(s)
    }
    syscall.ForkLock.RUnlock()
    if err != nil {
        return -1, err
    }
    if err = syscall.SetNonblock(s, true); err != nil {
        syscall.Close(s)
        return -1, err
    }
    return s, nil
}

// NOTE: Taken from the Go source: src/net/sockopt_posix.go
// Boolean to int.
func boolint(b bool) int {
    if b {
        return 1
    }
    return 0
}

// NOTE: Taken from the Go source: src/net/tcpsock_posix.go
func tcpAddrFamily(a *net.TCPAddr) int {
    if a == nil || len(a.IP) <= net.IPv4len {
        return syscall.AF_INET
    }
    if a.IP.To4() != nil {
        return syscall.AF_INET
    }
    return syscall.AF_INET6
}

// NOTE: Taken from the Go source: src/net/ipsock_posix.go
func ipToSockaddr(family int, sip string, port int) (syscall.Sockaddr, error) {
    switch family {
    case syscall.AF_INET:
		ip := net.ParseIP(sip)
        sa := new(syscall.SockaddrInet4)
        for i := 0; i < net.IPv4len; i++ {
            sa.Addr[i] = ip[i]
        }
        sa.Port = port
        return sa, nil
    }
    return nil, net.InvalidAddrError("unexpected socket family")
}

func Recv(c net.Conn, id int) {
	cache := make([]byte, 4096)
	fmt.Println("start thread:", id)
	for {
        n, err := c.Read(cache)
		fmt.Println("id:", id, " recv len:", n, " err:", err)
//		time.Sleep(time.Second / 10)
	}

}

func Recv2(c *net.UDPConn, id int) {
	cache := make([]byte, 4096)
	fmt.Println("start thread:", id)
	for {
        n, addr, err := c.ReadFromUDP(cache)
		fmt.Println("id:", id, " recv len:", n, " err:", err, "addr:", addr)
//		time.Sleep(time.Second / 10)
	}
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU());
	for i := 0; i < 20; i++ {
		c, err := openTransparent2("0.0.0.0", 3333)
		if err != nil {
			fmt.Println("create sock err:", err)
			os.Exit(-1)
		}
		go Recv2(c, i)
	}
	time.Sleep(time.Hour)
}



