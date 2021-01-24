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
	verbose = flag.Bool("v", false, "verbose output for debugging")
	addr    = flag.String("a", "localhost:9999", "address to listen on")
)

func usage() {
	fmt.Fprintf(flag.CommandLine.Output(),
		"USAGE: %s [FLAGS] MPV_SOCKET\n\n"+
			"The following flags are supported:\n\n", os.Args[0])

	flag.PrintDefaults()
	os.Exit(2)
}

func handleError(pc <-chan error, sc <-chan error) {
	for {
		select {
		case perr := <-pc:
			log.Println("[player error]", perr)
		case serr := <-sc:
			log.Println("[state error]", serr)
		}
	}
}

func startServer(mpvClient *mpv.Client, state *playerState) {
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}

	config := playlistfs.Config{
		&playctl{state: state, mpv: mpvClient},
		&playlist{state: state, mpv: mpvClient},
		&playvol{eofAt: -1, state: state, mpv: mpvClient},
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
	state, err := newPlayerState(mpvClient)
	if err != nil {
		log.Fatal(err)
	}

	// Use mpv debug mode if verbose output was requested
	mpvClient.Debug = *verbose

	go handleError(mpvClient.ErrChan, state.ErrChan())
	startServer(mpvClient, state)
}
