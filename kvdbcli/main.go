package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", ":9003")
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	conn.Write([]byte("BEGIN\r\nSET key1 value1\r\nGET key1\r\nDEL key1\r\nCOMMIT\r\nGET key1\r\nQUIT\r\n"))
	buff := make([]byte, 1024*60)
	n, err := conn.Read(buff)
	for err == nil {
		fmt.Println(string(buff[:n]))
		n, err = conn.Read(buff)
	}
}
