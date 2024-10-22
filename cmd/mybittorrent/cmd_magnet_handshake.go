package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"strings"

	bencode "github.com/jackpal/bencode-go"
)

func executeMagnetHandshake(args []string) {
	if len(args) < 1 {
		log.Fatalln("usage: magnet_info <magnet-link>")
	}

	magnet, err := parseMagnet(args[0])
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
	fmt.Printf("Peer ID: %x\n", receivedBytes[48:68])

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

	requestPacket := PeerMessage{
		length:  uint32(extensionBuffer.Len()),
		id:      20,
		payload: extensionBuffer.Bytes(),
	}

	_, err = client.Write(requestPacket.serialize())
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

	extensionId := decodedExtensionResponse.(map[string]any)["m"].(map[string]any)["ut_metadata"]
	fmt.Printf("Peer Metadata Extension ID: %d\n", extensionId)
}
