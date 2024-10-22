package main

import (
	"log"
	"os"
)

func main() {
	command := os.Args[1]
	args := os.Args[2:]

	if command == "decode" {
		executeDecode(args)
	} else if command == "info" {
		executeInfo(args)
	} else if command == "peers" {
		executePeers(args)
	} else if command == "handshake" {
		executeHandshake(args)
	} else if command == "download_piece" {
		executeDownloadPiece(args)
	} else if command == "download" {
		executeDownload(args)
	} else if command == "magnet_parse" {
		executeMagnetParse(args)
	} else if command == "magnet_handshake" {
		executeMagnetHandshake(args)
	} else if command == "magnet_info" {
		executeMagnetInfo(args)
	} else if command == "magnet_download_piece" {
		executeMagnetDownloadPiece(args)
	} else if command == "magnet_download" {
		executeMagnetDownload(args)
	} else {
		log.Fatalf("Unknown command: %s\n", command)
	}
}
