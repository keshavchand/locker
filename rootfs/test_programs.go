package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

type Tests func() error

func Output() error {
	log.Println("OUTPUT: Running as child process")
	pid := os.Getpid()
	log.Printf("Running as child process with pid: %d\n", pid)
	return nil
}

func SocketConnect() error {
	log.Println("SOCKET CONNECT: Connecting to google.com")
	socket, err := net.Dial("tcp", "142.250.182.68:80")
	if err != nil {
		return err
	}
	defer socket.Close()

	fmt.Fprintf(socket, "GET / HTTP/1.0\r\n\r\n")
	written, err := io.Copy(io.Discard, socket)
	if err != nil {
		return err
	}

	log.Printf("Written %d bytes\n", written)
	return nil
}

func DnsResolver() error {
	log.Println("DNS LOOKUP: IP address of google.com")
	_, err := net.LookupIP("google.com")
	if err != nil {
		return err
	}
	return nil
}

func DnsResolverCustom() error {
	log.Println("DNS LOOKUP: IP address of google.com using custom resolver")
	resolver := net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: 5 * time.Second,
			}
			return d.DialContext(ctx, "udp", "8.8.8.8:53")
		},
	}
	ips, err := resolver.LookupHost(context.Background(), "www.google.com")
	if err != nil {
		return err
	}

	log.Println("IP address: ")
	for _, ip := range ips {
		log.Println(" -> :", ip)
	}
	return nil
}

func Ps() error {
	log.Println("PS: List all processes")
	dirs, err := os.ReadDir("/proc")
	if err != nil {
		return err
	}

	isNumber := func(s string) bool {
		for _, c := range s {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	}

	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}
		if !isNumber(dir.Name()) {
			continue
		}

		cmd := "/proc/" + dir.Name() + "/cmdline"
		data, err := os.ReadFile(cmd)
		if err != nil {
			log.Println("ERROR:", err)
		}
		log.Println(dir.Name(), ":", string(data))
	}

	return nil
}

func main() {
	log.SetFlags(log.Lshortfile)
	tests := []Tests{
		Output,
		SocketConnect,
		DnsResolverCustom,
		DnsResolver,
		Ps,
	}
	for _, test := range tests {
		if err := test(); err != nil {
			log.Println("ERROR:", err)
		}
	}
}
