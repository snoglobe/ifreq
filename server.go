package main

import (
	"encoding/binary"
	"encoding/json"
	"github.com/mkb218/gosndfile/sndfile"
	"log"
	"net"
	"strconv"
)

var info StationInfo

const sampleRate = 32000
const seconds = 1

func BeginStream(file chan *sndfile.File) {
	buffer := make([]int32, sampleRate*seconds)

	log.Println("starting stream on port " + strconv.Itoa(config.Port()+1))

	listen, err := net.Listen("tcp", config.Bind().String()+":"+strconv.Itoa(config.Port()+1))
	if err != nil {
		panic(err)
	}
	defer listen.Close()

	go func() {
		for {
			select {
			case f := <-file:
				_, err := f.ReadFrames(buffer)
				if err != nil {
					break
				}
			}
		}
	}()

	for {
		conn, err := listen.Accept()
		if err != nil {
			panic(err)
		}
		go func(conn net.Conn) {
			for {
				binary.Write(conn, binary.LittleEndian, &buffer)
			}
		}(conn)
	}
}

func HandleConnection(conn net.Conn) {
	defer conn.Close()

	var req = Request{}
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		return
	}

	switch req.Type() {
	case RetrieveInfo:
		println(info.String())
		data, _ := json.Marshal(&Response{
			payload: &info,
			success: true,
		})
		println(string(data))
		_, err := conn.Write(data)
		if err != nil {
			return
		}
	case RequestStream:
		data, _ := json.Marshal(&Response{
			payload: config.Bind().String() + ":" + strconv.Itoa(config.Port()+1),
			success: true,
		})
		_, err := conn.Write(data)
		if err != nil {
			return
		}
	case RequestPeers:
		data, _ := json.Marshal(Response{
			payload: peers,
			success: true,
		})
		_, err := conn.Write(data)
		if err != nil {
			return
		}
	case GoingOffAir:
		for i, peer := range peers {
			station := req.Payload().(Station)
			if peer.Equal(&station) {
				peers = append(peers[:i], peers[i+1:]...)
			}
		}
		data, _ := json.Marshal(Response{
			payload: nil,
			success: true,
		})
		_, err := conn.Write(data)
		if err != nil {
			return
		}
	case GoingOnAir:
		// if peer not already in peers
		for _, peer := range peers {
			station := req.Payload().(Station)
			if peer.Equal(&station) {
				data, _ := json.Marshal(Response{
					payload: nil,
					success: true,
				})
				_, err := conn.Write(data)
				if err != nil {
					return
				}
				return
			}
		}
		peers = append(peers, Station{
			ip:   conn.RemoteAddr().(*net.TCPAddr).IP,
			port: uint16(conn.RemoteAddr().(*net.TCPAddr).Port),
		})
		data, _ := json.Marshal(Response{
			payload: nil,
			success: true,
		})
		_, err := conn.Write(data)
		if err != nil {
			return
		}
	}
}
