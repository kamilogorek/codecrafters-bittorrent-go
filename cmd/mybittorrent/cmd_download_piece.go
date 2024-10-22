package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
)

type PeerMessage struct {
	length  uint32
	id      byte
	payload []byte
}

func (pm *PeerMessage) serialize() []byte {
	var payload bytes.Buffer
	var length [4]byte
	// 1 bit to account for message id
	binary.BigEndian.PutUint32(length[:], pm.length+1)
	payload.Write(length[:])
	payload.WriteByte(pm.id)
	payload.Write(pm.payload)
	return payload.Bytes()
}

func deserializePeerMessage(bytes []byte) PeerMessage {
	length := binary.BigEndian.Uint32(bytes[0:4])
	id := bytes[4]
	var payload []byte
	if length > 0 {
		payload = bytes[5:]
	}
	return PeerMessage{
		length:  length,
		id:      id,
		payload: payload,
	}
}

func executeDownloadPiece(args []string) {
	if len(args) < 4 {
		log.Fatalln("usage: download_piece -o <out_dir> <torrent_file> <piece>")
	}

	pieceIndex, err := strconv.Atoi(args[3])
	if err != nil {
		log.Fatalln("invalid piece number", err)
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

	pieceData, err := downloadPiece(client, DownloadPieceInfo{
		index:      pieceIndex,
		length:     metainfo.Info.PieceLength,
		fileLength: metainfo.Info.Length,
	})
	if err != nil {
		log.Fatalln("Unable to download piece", pieceIndex, err)
	}

	sum := sha1.Sum(pieceData)
	if !slices.Equal(metainfo.Info.Pieces[pieceIndex], sum[:]) {
		log.Fatalf("\ninvalid hash of downloaded piece. want: %x, got: %x", metainfo.Info.Pieces[pieceIndex], sum[:])
	} else {
		fmt.Println("\nBlock checksum matches")
	}

	os.WriteFile(args[1], pieceData, os.ModePerm)
}
