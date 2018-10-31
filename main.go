package main

import (
	"bufio"
	"bytes"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	db = make(map[string]string)
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	wg := sync.WaitGroup{}

	go func() {
		<-sigs
		wg.Wait()
		l.Close()
		os.Exit(0)
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("err:", err)
		}
		wg.Add(1)
		go handleConn(conn, &wg)
	}
}

var dbLock sync.RWMutex
var db map[string]string

func handleConn(conn net.Conn, wg *sync.WaitGroup) {
	defer conn.Close()
	defer wg.Done()
	log.Println("debug: Connected")
	r := bufio.NewScanner(conn)
	r.Split(ScanLines)
	r.Buffer(make([]byte, 4096), bufio.MaxScanTokenSize)
	var out bytes.Buffer
	for r.Scan() {
		vals := strings.SplitN(r.Text(), " ", 2)
		out.Reset()
		switch vals[0] {
		case "GET":
			dbLock.RLock()
			val := db[vals[1]]
			dbLock.RUnlock()
			out.WriteString(val)
			out.WriteString("\n")
			conn.Write(out.Bytes())
		case "SET":
			vals = strings.SplitN(vals[1], " ", 2)
			dbLock.Lock()
			db[vals[0]] = vals[1]
			dbLock.Unlock()
			conn.Write([]byte("OK\n"))
		case "DEL":
			dbLock.Lock()
			delete(db, vals[1])
			dbLock.Unlock()
			conn.Write([]byte("OK\n"))
		case "QUIT":
			break
		case "BEGIN":
			if handleTransaction(r, conn) {
				break
			}
		case "COMMIT":
			conn.Write([]byte("not implemented\r\n"))
		default:
			conn.Write([]byte("not valid command\r\n"))
		}
	}
	log.Println("debug: Disconnected")
}

func handleTransaction(r *bufio.Scanner, conn net.Conn) bool {
	tmp := make(map[string]string)
	deleted := make(map[string]bool)
	dbLock.Lock()
	for k, v := range db {
		tmp[k] = v
	}
	dbLock.Unlock()
	var out bytes.Buffer
	for r.Scan() {
		out.Reset()
		vals := strings.SplitN(r.Text(), " ", 2)
		// fmt.Println(vals)
		switch vals[0] {
		case "GET":
			out.WriteString(tmp[vals[1]])
			out.WriteString("\n")
			//fmt.Printf("GET: %s [%s]\n", vals[1], string(out.Bytes()))
			conn.Write(out.Bytes())
		case "SET":
			vals = strings.SplitN(vals[1], " ", 2)
			tmp[vals[0]] = vals[1]
			//fmt.Printf("SET: [%s] [%s]\n", vals[0], tmp[vals[0]])
			delete(deleted, vals[0])
			conn.Write([]byte("OK\n"))
		case "DEL":
			delete(tmp, vals[1])
			deleted[vals[1]] = true
			conn.Write([]byte("OK\n"))
		case "COMMIT":
			dbLock.Lock()
			for k, v := range tmp {
				db[k] = v
			}
			for k, v := range deleted {
				if v {
					delete(db, k)
				}
			}
			dbLock.Unlock()
			//		conn.Write([]byte("OK\r\n"))
		case "QUIT":
			return true
		}
	}
	return false
}

func ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte{'\r', '\n'}); i >= 0 {
		return i + 2, data[0:i], nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), data, nil
	}
	// Request more data.
	return 0, nil, nil
}
