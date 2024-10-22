package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"strings"

	bencode "github.com/jackpal/bencode-go"
)

func executeMagnetDownload(args []string) {
	if len(args) < 3 {
		log.Fatalln("usage: magnet_download -o <out_dir> <magnet-link>")
	}

	magnet, err := parseMagnet(args[2])
	if err != nil {
		log.Fatalln("error parsing magnet link", err)
	}

	metainfo := Metainfo{
		Announce: magnet.tracker_url,
		InfoHash: []byte(magnet.info_hash),
		Info: Info{
			Name:   magnet.file_name,
			Length: 42,
		},
	}

	peers, err := getPeersFromMetainfo(&metainfo)
	if err != nil {
		log.Fatalln("unable to get peers from metainfo", err)
	}

	client, err := NewTCPClient(peers[0])
	if err != nil {
		log.Fatalln("unable to create tcp client", err)
	}

	handshakePacket := createHandshakePacket(metainfo.InfoHash, true)
	_, err = client.Write(handshakePacket)
	if err != nil {
		log.Fatalln("unable to write all data to client", err)
	}

	receivedBytes, err := client.ReceiveExact(68)
	if err != nil {
		log.Fatalln("unable to read data from client", err)
	}

	// ignore bitfield msg
	_, err = client.ReceiveExact(6)
	if err != nil {
		log.Fatalln("unable to read data from client", err)
	}

	extensionPayload := map[string]any{
		"m": map[string]any{
			"ut_metadata": 42,
		},
	}
	encodedExtensionPayload := new(strings.Builder)
	bencode.Marshal(encodedExtensionPayload, extensionPayload)

	extensionBuffer := new(bytes.Buffer)
	binary.Write(extensionBuffer, binary.BigEndian, []byte{0})
	binary.Write(extensionBuffer, binary.BigEndian, []byte(encodedExtensionPayload.String()))

	extensionHandshake := PeerMessage{
		length:  uint32(extensionBuffer.Len()),
		id:      20,
		payload: extensionBuffer.Bytes(),
	}

	_, err = client.Write(extensionHandshake.serialize())
	if err != nil {
		log.Fatalln("unable to write all data to socket", err)
	}

	_, receivedBytes, err = client.Receive(1024)
	if err != nil {
		log.Fatalln("unable to read data from socket")
	}

	decodedExtensionResponse, err := decodeBencode(string(receivedBytes[6:]))
	if err != nil {
		log.Fatalln(err)
	}

	extensionId := decodedExtensionResponse.(map[string]any)["m"].(map[string]any)["ut_metadata"].(int)

	requestPayload := map[string]any{
		"msg_type": 0,
		"piece":    0,
	}
	encodedRequestPayload := new(strings.Builder)
	bencode.Marshal(encodedRequestPayload, requestPayload)

	requestBuffer := new(bytes.Buffer)
	binary.Write(requestBuffer, binary.BigEndian, byte(extensionId))
	binary.Write(requestBuffer, binary.BigEndian, []byte(encodedRequestPayload.String()))

	requestPacket := PeerMessage{
		length:  uint32(requestBuffer.Len()),
		id:      20,
		payload: requestBuffer.Bytes(),
	}

	_, err = client.Write(requestPacket.serialize())
	if err != nil {
		log.Fatalln("unable to write all data to socket", err)
	}

	_, receivedBytes, err = client.Receive(1024)
	if err != nil {
		log.Fatalln("unable to read data from socket")
	}

	decodedRequestResponse, err := decodeBencode(string(receivedBytes[6:]))
	if err != nil {
		log.Fatalln(err)
	}

	firstDict := new(strings.Builder)
	bencode.Marshal(firstDict, decodedRequestResponse)
	decodedPiecesDict, err := decodeBencode(string(receivedBytes[firstDict.Len()+6:]))
	if err != nil {
		log.Fatalln(err)
	}

	rawPieces := []byte(decodedPiecesDict.(map[string]any)["pieces"].(string))
	piecesCount := len(rawPieces) / 20
	pieces := make([][]byte, 0, piecesCount)
	for i := 0; i < piecesCount; i++ {
		pieces = append(pieces, rawPieces[i*20:(i+1)*20])
	}

	fmt.Printf("Tracker URL: %s\n", metainfo.Announce)
	fmt.Printf("Length: %d\n", decodedPiecesDict.(map[string]any)["length"])
	fmt.Printf("Info Hash: %x\n", metainfo.InfoHash)
	fmt.Printf("Piece Length: %d\n", decodedPiecesDict.(map[string]any)["piece length"])
	fmt.Printf("Piece Hashes:\n")
	for _, piece := range pieces {
		fmt.Printf("%x\n", piece)
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

	_, receivedBytes, err = client.Receive(1024)
	if err != nil {
		log.Fatalln("unable to read data from socket")
	}

	err = downloadFile(client, DownloadFileInfo{
		path:        args[1],
		fileLength:  decodedPiecesDict.(map[string]any)["length"].(int),
		pieceLength: decodedPiecesDict.(map[string]any)["piece length"].(int),
		pieceHashes: pieces,
	})
	if err != nil {
		log.Fatalln("unable to download file", err)
	}
}
