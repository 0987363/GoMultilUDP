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



