package main

import (
	"fmt"
	"net"
	"os"
	"path"

	"github.com/npat-efault/varhacks/chanio"
)

func Usage(cmd string) {
	fmt.Fprintf(os.Stderr, "Usage is: %s <local addr>\n", cmd)
}

func main() {
	if len(os.Args) != 2 {
		Usage(path.Base(os.Args[0]))
		os.Exit(1)
	}
	ln, err := net.Listen("tcp", os.Args[1])
	if err != nil {
		fmt.Println("Listen:", err)
		os.Exit(1)
	}
	conn, err := ln.Accept()
	if err != nil {
		fmt.Println("Accept:", err)
		os.Exit(1)
	}

	rx := chanio.NewRx(conn, 128, nil)
	tx := chanio.NewTx(conn, nil)
	var rxp <-chan chanio.Packet = rx.Pck()
	var txp chan<- []byte = nil
	var p chanio.Packet
	for {
		select {
		case p = <-rxp:
			if p.Err != nil {
				fmt.Println("Rx:", p.Err)
				rx.Close()
				tx.Close()
				os.Exit(1)
			}
			rxp = nil
			txp = tx.Pck()
		case txp <- p.Pck:
			rxp = rx.Pck()
			txp = nil
		case err := <-tx.Err():
			fmt.Println("Tx:", err)
			rx.Close()
			tx.Close()
			os.Exit(1)
		}
	}
}
