package test

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"testing"
	"time"
)

func connectToService() interface{} {
	time.Sleep(1 * time.Second)
	return struct{}{}
}

func warmServiceConnCache() *sync.Pool {
	var p *sync.Pool

	p = &sync.Pool{
		New: connectToService,
	}

	for i := 0; i < 10; i++ {
		p.Put(p.New())
	}

	return p
}

func startNetworkDaemon() *sync.WaitGroup {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		server, err := net.Listen("tcp", "localhost:8080")
		if err != nil {
			log.Fatalf("cannot listen: %v", err)
		}
		defer server.Close()

		wg.Done()

		for {
			conn, err := server.Accept()
			if err != nil {
				log.Printf("cannot accept connection: %d", err)
				continue
			}

			connectToService()
			fmt.Println("")
			conn.Close()
		}
	}()

	return &wg
}

func startNetworkDaemon2() *sync.WaitGroup {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		conn_pool := warmServiceConnCache()

		server, err := net.Listen("tcp", "localhost:8080")
		if err != nil {
			log.Fatalf("cannot listen: %v", err)
		}
		defer server.Close()

		wg.Done()

		for {
			conn, err := server.Accept()
			if err != nil {
				log.Printf("cannot accept connection: %v", err)
				continue
			}

			svc_conn := conn_pool.Get()
			fmt.Fprintln(conn, "")
			conn_pool.Put(svc_conn)
			conn.Close()
		}
	}()

	return &wg
}

func init() {
	daemonStarted := startNetworkDaemon2()
	daemonStarted.Wait()
}

func BenchmarkNetworkRequest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		conn, err := net.Dial("tcp", "localhost:8080")
		if err != nil {
			b.Fatalf("cannot dial host: %v", err)
		}
		defer conn.Close()

		_, err = io.ReadAll(conn)
		if err != nil {
			b.Fatalf("cannot read: %v", err)
		}
	}
}
