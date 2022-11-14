package main

import (
	"github.com/mkb218/gosndfile/sndfile"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

var peers = make([]Station, 0)
var config *Config

func contains(s []Station, e Station) bool {
	for _, v := range s {
		if v.Equal(&e) {
			return true
		}
	}
	return false
}

func main() {
	// load peers from config file
	var err error
	config, err = LoadConfig("config.json")
	if err != nil {
		log.Fatal(err)
	}

	// contact peers and see if they have more peers to share
	peers = config.Peers()
	for _, peer := range peers {
		peerContact := NewRequest(RequestPeers, nil)
		peerResponse := SendRequest(peerContact, peer.IP(), peer.Port())
		if !peerResponse.Success() {
			config.RemovePeer(&peer)
			continue
		}
		peerPeers := peerResponse.Payload().([]Station)
		for _, peerPeer := range peerPeers {
			if !contains(peers, peerPeer) {
				peers = append(peers, peerPeer)
			}
		}
	}

	defer config.Save("config.json")

	// server or client mode
	switch os.Args[1] {
	case "server":
		server()
	case "client":
		endChan := make(chan int)
		go ConnectPeer(&Station{
			ip:   net.ParseIP(os.Args[2]),
			port: 830,
		}, endChan)
		time.Sleep(100 * time.Second)
		endChan <- 1
	default:
		log.Fatal("invalid argument")
	}
}

func server() {
	defer SaveStationConfig("station.json", info)
	// check if files directory exists
	if _, err := os.Stat("files"); os.IsNotExist(err) {
		log.Fatal("files directory does not exist")
	}

	// load station config
	station, err := LoadStationConfig("station.json")
	info = station

	// start tcp server
	log.Println("starting server on port " + strconv.Itoa(config.Port()))
	ln, err := net.Listen("tcp", config.Bind().String()+":"+strconv.Itoa(config.Port()))

	if err != nil {
		log.Fatal(err)
	}

	files := make(chan *sndfile.File)
	list, err := os.ReadDir("files")
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range list {
		go func(file os.DirEntry) {
			var i sndfile.Info
			f, err := sndfile.Open("files/"+file.Name(), sndfile.Read, &i)
			if i.Samplerate != sampleRate && i.Format != sndfile.SF_FORMAT_PCM_32 {
				log.Println("invalid file, skipping: " + file.Name())
				return
			}
			if err != nil {
				log.Fatal(err)
			}
			files <- f
		}(file)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go HandleConnection(conn)
	}
}
