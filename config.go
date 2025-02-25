package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	ElevatorAddr string `json:"elevator_addr"`
	NumFloors    int    `json:"num_floors"`
	LocalPeerId  int    `json:"local_peer_id"`
	LocalAddr    string `json:"local_addr"`
	LocalPort    uint16 `json:"local_port"`
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &Config{}
	err = decoder.Decode(config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
