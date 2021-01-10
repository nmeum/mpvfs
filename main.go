package main

import (
	"github.com/nmeum/mpvfs/mpv"
	"github.com/nmeum/mpvfs/playlistfs"
	"go.rbn.im/neinp"

	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

var (
	addr = flag.String("a", "localhost:9999", "address to listen on")
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(),
		"USAGE: %s [FLAGS] MPV_SOCKET\n\n"+
			"The following flags are supported:\n\n", os.Args[0])

	flag.PrintDefaults()
	os.Exit(2)
}

func startServer(mpvClient *mpv.Client) {
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	config := playlistfs.Config{
		playctl{mpvClient},
		ControlFile{"playlist"},
		ControlFile{"playvol"},
	}

	fs := playlistfs.NewPlaylistFS(config)
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

func main() {
	flag.Parse()
	flag.Usage = usage

	if flag.NArg() < 1 {
		usage()
	}
	socketFp := flag.Arg(0)

	mpvClient, err := mpv.NewClient(socketFp)
	if err != nil {
		log.Fatal(err)
	}
	go func(ch <-chan error) {
		err := <-ch
		log.Println(err)
	}(mpvClient.ErrChan)

	startServer(mpvClient)
}
