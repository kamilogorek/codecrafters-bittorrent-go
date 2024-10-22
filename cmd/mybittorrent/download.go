package main

import (
	"bytes"
	"cmp"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"os"
	"slices"
)

type DownloadFileInfo struct {
	path        string
	fileLength  int
	pieceLength int
	pieceHashes [][]byte
}

type DownloadPieceInfo struct {
	index      int
	length     int
	fileLength int
}

type DownloadBlockInfo struct {
	index       int
	count       int
	size        int
	pieceLength int
	pieceIndex  int
}

type BlockRequest struct {
	index  uint32
	begin  uint32
	length uint32
}

func downloadBlock(client *TCPClient, info DownloadBlockInfo) ([]byte, error) {
	fmt.Printf("\nBlock index: %d/%d\n", info.index, info.count-1)

	begin := uint32(info.index * info.size)
	length := uint32(info.size)
	if info.index == info.count-1 {
		length = uint32(cmp.Or(info.pieceLength%info.size, info.size))
		fmt.Printf("Block size (last): %d\n", length)
	} else {
		fmt.Printf("Block size: %d\n", length)
	}

	blockBuffer := new(bytes.Buffer)
	blockRequest := BlockRequest{
		index:  uint32(info.pieceIndex),
		begin:  begin,
		length: length,
	}
	binary.Write(blockBuffer, binary.BigEndian, blockRequest)
	fmt.Printf("[req] block: %+v\n", blockRequest)

	requestPacket := PeerMessage{
		length:  uint32(blockBuffer.Len()),
		id:      6,
		payload: blockBuffer.Bytes(),
	}
	_, err := client.Write(requestPacket.serialize())
	if err != nil {
		return nil, fmt.Errorf("unable to write all data to socket: %w", err)
	}

	receivedBytes, err := client.ReceiveExact(4)
	if err != nil {
		return nil, fmt.Errorf("unable to read data from socket: %w", err)
	}
	blockPacketLen := binary.BigEndian.Uint32(receivedBytes)

	fmt.Printf("Next block len: %d\n", blockPacketLen)

	receivedBytes, err = client.ReceiveExact(int(blockPacketLen))
	if err != nil {
		return nil, fmt.Errorf("unable to read data from socket: %w", err)
	}

	fmt.Printf("[resp] id: %d, index: %d, begin: %d, data_len: %d\n",
		receivedBytes[0],
		binary.BigEndian.Uint32(receivedBytes[1:5]),
		binary.BigEndian.Uint32(receivedBytes[5:9]),
		len(receivedBytes[9:]),
	)

	return receivedBytes[9:], nil
}

func downloadPiece(client *TCPClient, info DownloadPieceInfo) ([]byte, error) {
	pieceLength := info.length
	count := int(math.Ceil(float64(info.fileLength) / float64(pieceLength)))
	if info.index == count-1 {
		pieceLength = cmp.Or(info.fileLength%info.length, info.length)
	}

	fmt.Printf("\nDownloading piece %d/%d\n", info.index, count-1)

	blockSize := 16 * 1024
	blockCount := int(math.Ceil(float64(pieceLength) / float64(blockSize)))

	fmt.Printf("\nLength: %d\n", info.fileLength)
	fmt.Printf("Piece count: %d\n", count)
	fmt.Printf("Piece number: %d/%d\n", info.index, count-1)
	fmt.Printf("Max piece length: %d\n", info.length)
	fmt.Printf("Current piece length: %d\n", pieceLength)
	fmt.Printf("Block count: %d\n", blockCount)

	var pieceData []byte
	for blockIndex := 0; blockIndex < blockCount; blockIndex++ {
		fmt.Printf("Filesize: %d/%d\n", len(pieceData), pieceLength)
		blockData, err := downloadBlock(client, DownloadBlockInfo{
			index:       blockIndex,
			count:       blockCount,
			size:        blockSize,
			pieceLength: pieceLength,
			pieceIndex:  info.index,
		})
		if err != nil {
			log.Fatalln(err)
		}
		pieceData = append(pieceData, blockData...)
	}
	return pieceData, nil
}

func downloadFile(client *TCPClient, info DownloadFileInfo) error {
	pieceCount := int(math.Ceil(float64(info.fileLength) / float64(info.pieceLength)))

	f, err := os.Create(info.path)
	if err != nil {
		log.Fatalln("cannot create file", err)
	}
	defer f.Close()

	for pieceIndex := 0; pieceIndex < pieceCount; pieceIndex++ {
		pieceData, err := downloadPiece(client, DownloadPieceInfo{
			index:      pieceIndex,
			length:     info.pieceLength,
			fileLength: info.fileLength,
		})
		if err != nil {
			log.Fatalln("Unable to download piece", pieceIndex, err)
		}

		sum := sha1.Sum(pieceData)
		if !slices.Equal(info.pieceHashes[pieceIndex], sum[:]) {
			log.Fatalf("Invalid hash of downloaded piece. want: %x, got: %x", info.pieceHashes[pieceIndex], sum[:])
		} else {
			fmt.Println("\nBlock checksum matches")
		}

		f.Write(pieceData)
	}

	return nil
}
