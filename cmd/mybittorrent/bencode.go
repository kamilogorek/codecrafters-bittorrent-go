package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"unicode"

	bencode "github.com/jackpal/bencode-go"
)

func decodeBencode(input string) (any, error) {
	value, _, err := decodeBencodeImpl(input)
	return value, err
}

func decodeBencodeImpl(input string) (any, string, error) {
	if unicode.IsDigit(rune(input[0])) {
		var firstColonIndex int

		for i := 0; i < len(input); i++ {
			if input[i] == ':' {
				firstColonIndex = i
				break
			}
		}

		lengthStr := input[:firstColonIndex]
		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return nil, "", err
		}

		value := input[firstColonIndex+1 : firstColonIndex+1+length]
		rem := input[len(lengthStr)+length+1:]

		return value, rem, nil
	} else if input[0] == 'i' {
		var firstEndIndex int

		for i := 0; i < len(input); i++ {
			if input[i] == 'e' {
				firstEndIndex = i
				break
			}
		}

		value, err := strconv.ParseInt(input[1:firstEndIndex], 10, 64)
		if err != nil {
			return nil, "", err
		}
		rem := input[firstEndIndex+1:]

		return int(value), rem, err
	} else if input[0] == 'l' {
		values := make([]any, 0)
		rest := input[1:]

		for len(rest) > 1 {
			value, rem, err := decodeBencodeImpl(rest)
			if err != nil {
				return nil, rest, err
			}
			values = append(values, value)
			if rem[0] == 'e' {
				rest = rem[1:]
				break
			} else {
				rest = rem
			}
		}

		return values, rest, nil
	} else if input[0] == 'd' {
		dict := make(map[string]any, 0)
		rest := input[1:]

		for len(rest) > 1 {
			key, rem, err := decodeBencodeImpl(rest)
			value, rem, err := decodeBencodeImpl(rem)
			if err != nil {
				return nil, rest, err
			}
			dict[key.(string)] = value
			if rem[0] == 'e' {
				rest = rem[1:]
				break
			} else {
				rest = rem
			}
		}

		return dict, rest, nil
	} else {
		return nil, "", fmt.Errorf("Unknown input: %s", input)
	}
}

type Metainfo struct {
	Announce string
	Info     Info
	InfoHash []byte
}

type Info struct {
	Length      int
	Name        string
	PieceLength int
	Pieces      [][]byte
}

func parseMetainfo(r io.Reader) (Metainfo, error) {
	input, err := io.ReadAll(r)
	if err != nil {
		return Metainfo{}, fmt.Errorf("unable to read metainfo file contents")
	}

	decoded, err := decodeBencode(string(input))
	if err != nil {
		log.Fatalln(err)
	}

	decodedMetainfo := decoded.(map[string]any)
	decodedInfo := decodedMetainfo["info"].(map[string]any)
	announce := decodedMetainfo["announce"].(string)
	length := decodedInfo["length"].(int)
	name := decodedInfo["name"].(string)
	pieceLength := decodedInfo["piece length"].(int)
	rawPieces := []byte(decodedInfo["pieces"].(string))
	piecesCount := len(rawPieces) / 20
	pieces := make([][]byte, 0, piecesCount)
	for i := 0; i < piecesCount; i++ {
		pieces = append(pieces, rawPieces[i*20:(i+1)*20])
	}
	out := new(strings.Builder)
	bencode.Marshal(out, decodedInfo)
	infoHash := sha1.Sum([]byte(out.String()))

	return Metainfo{
		Announce: announce,
		Info: Info{
			Length:      length,
			Name:        name,
			PieceLength: pieceLength,
			Pieces:      pieces,
		},
		InfoHash: infoHash[:],
	}, nil
}
