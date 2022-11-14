package main

import (
	"encoding/json"
	"net"
	"net/url"
	"os"
)

type Config struct {
	peers          []Station
	defaultStation *Station
	port           int
	bind           net.IP
}

func (c *Config) Peers() []Station {
	return c.peers
}

func (c *Config) DefaultStation() *Station {
	return c.defaultStation
}

func (c *Config) Port() int {
	return c.port
}

func (c *Config) Bind() net.IP {
	return c.bind
}

func (c *Config) SetDefaultStation(station *Station) {
	c.defaultStation = station
}

func (c *Config) AddPeer(station *Station) {
	c.peers = append(c.peers, *station)
}

func (c *Config) RemovePeer(station *Station) {
	for i, peer := range c.peers {
		if peer.Equal(station) {
			c.peers = append(c.peers[:i], c.peers[i+1:]...)
		}
	}
}

func (c *Config) Save(file string) error {
	config, err := json.Marshal(map[string]any{
		"peers":          c.peers,
		"defaultStation": c.defaultStation,
		"port":           c.port,
		"bind":           c.bind,
	})
	if err != nil {
		return err
	}
	err = os.WriteFile(file, config, 0644)
	if err != nil {
		return err
	}
	return nil
}

func MakeIfNotExist(file string) (bool, error) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		create, err := os.Create(file)
		defer create.Close()
		if err != nil {
			return true, err
		}
		return true, nil
	}
	return false, nil
}

func LoadConfig(file string) (*Config, error) {
	if made, err := MakeIfNotExist(file); err == nil && made {
		defaultConfig, err := json.Marshal(map[string]any{
			"peers":          []Station{},
			"defaultStation": nil,
			"port":           830,
			"bind":           net.IP{0, 0, 0, 0},
		})
		if err != nil {
			return nil, err
		}
		err = os.WriteFile(file, defaultConfig, 0644)
		if err != nil {
			return nil, err
		}
		return &Config{[]Station{}, nil, 830, net.IP{0, 0, 0, 0}}, nil
	} else if err != nil {
		return nil, err
	}
	configFile, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var config map[string]any
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}

	peers := make([]Station, len(config["peers"].([]any)))
	for i, peer := range config["peers"].([]any) {
		peers[i] = peer.(Station)
	}
	var defaultStation *Station
	if config["defaultStation"] == nil {
		defaultStation = nil
	} else {
		defaultStation = config["defaultStation"].(*Station)
	}
	port := int(config["port"].(float64))
	bind := net.ParseIP(config["bind"].(string))

	return &Config{peers, defaultStation, port, bind}, nil
}

func LoadStationConfig(file string) (StationInfo, error) {
	if made, err := MakeIfNotExist(file); err == nil && made {
		defaultConfig, err := json.Marshal(map[string]any{
			"name": "Default",
			"media": &Media{
				mediaType: Talk,
				artist:    "You",
				title:     "Your Cool Show",
				album:     "",
				albumArt:  url.URL{},
			},
			"runningTime": 0,
		})
		if err != nil {
			return StationInfo{}, err
		}
		err = os.WriteFile(file, defaultConfig, 0644)
		if err != nil {
			return StationInfo{}, err
		}
		return StationInfo{
			"Default",
			&Media{
				mediaType: Talk,
				artist:    "You",
				title:     "Your Cool Show",
				album:     "",
				albumArt:  url.URL{},
			},
			0,
		}, nil
	}
	configFile, err := os.ReadFile(file)
	if err != nil {
		return StationInfo{}, err
	}

	var config map[string]any
	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return StationInfo{}, err
	}

	name := config["name"].(string)

	mediaType := MediaType(byte(int(config["media"].(map[string]any)["mediaType"].(float64))))
	artist := config["media"].(map[string]any)["artist"].(string)
	title := config["media"].(map[string]any)["title"].(string)
	album := config["media"].(map[string]any)["album"].(string)
	var albumArt url.URL
	urlData, err := json.Marshal(config["media"].(map[string]any)["albumArt"])
	err = json.Unmarshal(urlData, &albumArt)

	media := Media{
		mediaType: mediaType,
		artist:    artist,
		title:     title,
		album:     album,
		albumArt:  albumArt,
	}

	runningTime := uint32(config["runningTime"].(float64))

	return StationInfo{name, &media, runningTime}, nil
}

func SaveStationConfig(file string, station StationInfo) error {
	config, err := json.Marshal(map[string]any{
		"name":        station.Name,
		"media":       station.Media,
		"runningTime": station.RunningTime,
	})
	if err != nil {
		return err
	}
	err = os.WriteFile(file, config, 0644)
	if err != nil {
		return err
	}
	return nil
}
