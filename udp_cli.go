package main

import (
	"fmt"
	"net"
	"os"
	"time"
    "runtime"
)

func UdpSend(srvIP string, srcPort int) {
	addr := net.UDPAddr{
		Port: 3333,
		IP: net.ParseIP(srvIP),
	}
	/*
	laddr := net.UDPAddr{
		Port: srcPort,
		IP: net.ParseIP(srvIP),
	}
	*/
	conn, err := net.DialUDP("udp", nil, &addr)
	if (err != nil) {
		fmt.Println("udp connect failed, ", err)
		os.Exit(-1)
	}
	data := make([]byte, 100)

	for {
		conn.Write(data)
//		time.Sleep(time.Second)
	}
}

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU());
	for i := 0; i < 2; i++ {
		go UdpSend("192.168.1.62", 6666 + i)
		time.Sleep(time.Second / 10)
	}

	time.Sleep(time.Hour)
}




