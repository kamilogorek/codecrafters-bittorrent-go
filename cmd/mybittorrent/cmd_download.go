package main

import (
	"fmt"
	"log"
	"os"
)

func executeDownload(args []string) {
	if len(args) < 3 {
		log.Fatalln("usage: download -o <out_dir> <torrent_file>")
	}

	f, err := os.Open(args[2])
	if err != nil {
		log.Fatalln("error opening metainfo file", err)
	}
	defer f.Close()

	metainfo, err := parseMetainfo(f)
	if err != nil {
		log.Fatalln("error parsing metainfo file", err)
	}

	peers, err := getPeersFromMetainfo(&metainfo)
	if err != nil {
		log.Fatalln("unable to get peers from metainfo", err)
	}
	if len(peers) == 0 {
		log.Fatalln("no peers available")
	}

	client, err := NewTCPClient(peers[0])
	if err != nil {
		log.Fatalln("unable to create tcp client")
	}

	handshakePacket := createHandshakePacket(metainfo.InfoHash, false)
	_, err = client.Write(handshakePacket)
	if err != nil {
		log.Fatalln("unable to write all data to socket")
	}

	receivedLen, receivedBytes, err := client.Receive(1024)
	if err != nil {
		log.Fatalln("unable to read data from socket")
	}
	fmt.Printf("Peer ID: %x\n", receivedBytes[receivedLen-20:])

	receivedLen, receivedBytes, err = client.Receive(1024)
	if err != nil {
		log.Fatalln("unable to read data from socket")
	}

	interestedPacket := PeerMessage{
		length:  1,
		id:      2,
		payload: make([]byte, 0),
	}
	_, err = client.Write(interestedPacket.serialize())
	if err != nil {
		log.Fatalln("unable to write all data to socket")
	}

	receivedLen, receivedBytes, err = client.Receive(1024)
	if err != nil {
		log.Fatalln("unable to read data from socket")
	}

	err = downloadFile(client, DownloadFileInfo{
		path:        args[1],
		fileLength:  metainfo.Info.Length,
		pieceLength: metainfo.Info.PieceLength,
		pieceHashes: metainfo.Info.Pieces,
	})
	if err != nil {
		log.Fatalln("unable to download file", err)
	}

	fmt.Printf("File downloaded to %s\n", args[1])
}
