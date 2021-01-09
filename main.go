package main

import (
	"github.com/nmeum/mpvfs/playlistfs"
	"go.rbn.im/neinp"

	"flag"
	"io"
	"log"
	"net"
)

var (
	addr = flag.String("a", "localhost:9999", "address to listen on")
)

func main() {
	flag.Parse()

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	fs := playlistfs.NewPlaylistFS()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		server := neinp.NewServer(fs)

		err = server.Serve(conn)
		if err != nil && err != io.EOF {
			log.Println(err)
			continue
		}
	}

}
