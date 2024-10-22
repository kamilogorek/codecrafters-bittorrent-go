package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
)

type Magnet struct {
	info_hash   []byte
	file_name   string
	tracker_url string
}

func parseMagnet(link string) (Magnet, error) {
	magnet := Magnet{}

	if !strings.HasPrefix(link, "magnet:?") {
		return magnet, errors.New("invalid magnet link, missing magnet:? prefix")
	}

	linkValue := strings.TrimPrefix(link, "magnet:?")
	parts := strings.Split(linkValue, "&")
	values := make(map[string]string)

	for _, part := range parts {
		splitParts := strings.SplitN(part, "=", 2)
		key := splitParts[0]
		value := splitParts[1]
		values[key] = value
	}

	if !strings.HasPrefix(values["xt"], "urn:btih:") {
		return magnet, errors.New("invalid xt value, missing urn:btih: prefix")
	}

	xt, err := hex.DecodeString(values["xt"][9:])
	if err != nil {
		return magnet, errors.New("invalid xt hex value")
	}
	magnet.info_hash = xt
	magnet.file_name = values["dn"]
	unescaped, err := url.PathUnescape(values["tr"])
	if err != nil {
		return magnet, fmt.Errorf("invalid tracker_url value, cannot unescape the value: %s", values["tr"])
	}
	magnet.tracker_url = unescaped

	return magnet, nil
}

func executeMagnetParse(args []string) {
	if len(args) < 1 {
		log.Fatalln("usage: magnet_parse <magnet-link>")
	}

	magnet, err := parseMagnet(args[0])
	if err != nil {
		log.Fatalln("error parsing magnet link", err)
	}

	fmt.Printf("Tracker URL: %s\n", magnet.tracker_url)
	fmt.Printf("Info Hash: %x\n", magnet.info_hash)
}
