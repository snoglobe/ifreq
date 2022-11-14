package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
)

type Station struct {
	ip   net.IP
	port uint16
}

type MediaType byte

const (
	Music MediaType = iota
	News
	Weather
	Talk
)

type Media struct {
	mediaType MediaType

	artist   string
	title    string
	album    string
	albumArt url.URL
}

type StationInfo struct {
	name        string
	media       *Media
	runningTime uint32
}

func (s *StationInfo) Name() string {
	return s.name
}

func (s *StationInfo) Media() *Media {
	return s.media
}

func (s *StationInfo) RunningTime() uint32 {
	return s.runningTime
}

func (s *StationInfo) String() string {
	return fmt.Sprintf("%s: %s - %s | %s: %s", s.name, s.media.artist, s.media.title, s.media.album, s.media.albumArt.String())
}

func (s *Media) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"mediaType": s.mediaType,
		"artist":    s.artist,
		"title":     s.title,
		"album":     s.album,
		"albumArt":  s.albumArt,
	})
}

func (s *Media) MediaType() MediaType {
	return s.mediaType
}

func (s *Media) Artist() string {
	return s.artist
}

func (s *Media) Title() string {
	return s.title
}

func (s *Media) Album() string {
	return s.album
}

func (s *Media) AlbumArt() url.URL {
	return s.albumArt
}

func NewStation(ip net.IP, port uint16) *Station {
	return &Station{ip, port}
}

func (s *Station) IP() net.IP {
	return s.ip
}

func (s *Station) Port() uint16 {
	return s.port
}

func (s *Station) String() string {
	return fmt.Sprintf("%s:%d", s.ip.String(), s.port)
}

func (s *Station) StationInfo() StationInfo {
	req := SendRequest(NewRequest(RetrieveInfo, nil), s.ip, s.port)
	if req.Success() {
		return req.Payload().(StationInfo)
	}
	panic(req.payload.(string))
}

func (s *Station) Equal(other *Station) bool {
	return s.ip.Equal(other.ip) && s.port == other.port
}

func (s *Station) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"ip":   s.ip,
		"port": s.port,
	})
}

func (s *Station) UnmarshalJSON(data []byte) error {
	var station map[string]any
	if err := json.Unmarshal(data, &station); err != nil {
		return err
	}

	s.ip = station["ip"].(net.IP)
	s.port = uint16(station["port"].(float64))

	return nil
}
