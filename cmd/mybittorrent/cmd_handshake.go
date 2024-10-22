package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log"
	"os"
)

func createHandshakePacket(infohash []byte, isMagnet bool) []byte {
	var payload bytes.Buffer
	// length of the protocol string (BitTorrent protocol) which is 19 (1 byte)
	payload.WriteByte(19)
	// the string BitTorrent protocol (19 bytes)
	payload.WriteString("BitTorrent protocol")
	if isMagnet {
		payload.Write([]byte{0, 0, 0, 0, 0, 16, 0, 0})
	} else {
		// eight reserved bytes, which are all set to zero (8 bytes)
		payload.Write([]byte{0, 0, 0, 0, 0, 0, 0, 0})
	}
	// sha1 infohash (20 bytes) (NOT the hexadecimal representation, which is 40 bytes long)
	payload.Write(infohash)
	// peer id (20 bytes) (generate 20 random byte values)
	peerId := make([]byte, 20)
	rand.Read(peerId)
	payload.Write(peerId)
	return payload.Bytes()
}

func executeHandshake(args []string) {
	if len(args) < 2 {
		log.Fatalln("usage: handshake <torrent_file> <peer_addr>")
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

	client, err := NewTCPClient(args[1])
	if err != nil {
		log.Fatalln("unable to create tcp client", err)
	}

	handshakePacket := createHandshakePacket(metainfo.InfoHash, false)
	_, err = client.Write(handshakePacket)
	if err != nil {
		log.Fatalln("unable to write all data to client", err)
	}

	receivedLen, receivedBytes, err := client.Receive(68)
	if err != nil {
		log.Fatalln("unable to read data from client", err)
	}

	fmt.Printf("Peer ID: %x\n", receivedBytes[receivedLen-20:])
}
