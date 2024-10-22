package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strconv"
)

func getPeersFromMetainfo(metainfo *Metainfo) ([]string, error) {
	u, err := url.Parse(metainfo.Announce)
	if err != nil {
		log.Fatal("invalid announce url", err)
	}

	params := u.Query()
	params.Add("info_hash", string(metainfo.InfoHash))
	params.Add("peer_id", "dededededededededede")
	params.Add("port", "6881")
	params.Add("uploaded", "0")
	params.Add("downloaded", "0")
	params.Add("left", strconv.Itoa(metainfo.Info.Length))
	params.Add("compact", "1")
	u.RawQuery = params.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("unable to request peers", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read peers request body", err)
	}
	defer resp.Body.Close()

	decodedBody, err := decodeBencode(string(body))
	if err != nil {
		return nil, fmt.Errorf("unable to decode peers body %w", err)
	}
	rawPeers := []byte(decodedBody.(map[string]any)["peers"].(string))

	peersCount := len(rawPeers) / 6
	peers := make([]string, 0, peersCount)
	for i := 0; i < peersCount; i++ {
		rawPeer := rawPeers[i*6 : (i+1)*6]
		ip := netip.AddrFrom4([4]byte(rawPeer[0:5]))
		peer := netip.AddrPortFrom(ip, binary.BigEndian.Uint16(rawPeer[4:6]))
		peers = append(peers, peer.String())
	}
	return peers, nil
}

func executePeers(args []string) {
	if len(args) < 1 {
		log.Fatalln("usage: peers <torrent_file>")
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

	peers, err := getPeersFromMetainfo(&metainfo)
	if err != nil {
		log.Fatalln("unable to get peers from metainfo", err)
	}

	for _, v := range peers {
		fmt.Println(v)
	}
}
