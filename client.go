package main

import (
	"encoding/binary"
	"github.com/gordonklaus/portaudio"
	"log"
	"net"
	"strconv"
	"time"
)

func ConnectPeer(peer *Station, end chan int) {
	conn, err := net.Dial("tcp", peer.IP().String()+":"+strconv.Itoa(int(peer.Port())))
	if err != nil {
		log.Fatal(err)
	}
	// request info
	info := peer.StationInfo()
	log.Println("station info: " + info.String())
	// request stream
	getStream := NewRequest(RequestStream, nil)
	response := SendRequest(getStream, peer.IP(), peer.Port())
	if !response.Success() {
		log.Fatal("failed to get streamIpPort - " + response.Payload().(string))
	}
	streamIpPort := response.Payload().(string)
	conn.Close()

	conn, err = net.Dial("tcp", streamIpPort)
	if err != nil {
		log.Fatal(err)
	}

	buffer := make([]int32, sampleRate*seconds)

	// open portaudio output
	stream, err := portaudio.OpenDefaultStream(0, 2, 44100, 1024, func(out []int32) {
		// read from stream
		err := binary.Read(conn, binary.LittleEndian, &buffer)
		if err != nil {
			log.Fatal(err)
		}
		// write to portaudio
		for i := 0; i < len(buffer); i++ {
			out[i] = buffer[i]
		}
	})
	if err != nil {
		log.Fatal(err)
	}

	err = stream.Start()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		ticker := time.NewTicker(time.Second * 10)
		for {
			select {
			case <-ticker.C:
				info := peer.StationInfo()
				log.Println("station info: " + info.String())
			}
		}
	}()

	select {
	case <-end:
		stream.Stop()
		stream.Close()
		conn.Close()
	}
}
