package main

import (
	"fmt"
	"log"
	"os"
)

func executeInfo(args []string) {
	if len(args) < 1 {
		log.Fatalln("usage: info <torrent_file>")
	}

	f, err := os.Open(args[0])
	if err != nil {
		log.Fatalln("error opening metainfo file", err)
	}
	defer f.Close()

	metainfo, err := parseMetainfo(f)
	if err != nil {
		log.Fatalln("error parsing metainfo file", err)
	}

	fmt.Printf("Tracker URL: %s\n", metainfo.Announce)
	fmt.Printf("Length: %d\n", metainfo.Info.Length)
	fmt.Printf("Info Hash: %x\n", metainfo.InfoHash)
	fmt.Printf("Piece Length: %d\n", metainfo.Info.PieceLength)
	fmt.Println("Piece Hashes:")
	for _, piece := range metainfo.Info.Pieces {
		fmt.Printf("%x\n", piece)
	}
}
