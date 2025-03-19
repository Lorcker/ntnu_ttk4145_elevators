package main

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	// ElevatorAddr is the address of the elevator simulator or the elevator hardware.
	ElevatorAddr string `json:"elevator_addr"`

	// LocalPeerId is the id of the local elevator.
	LocalPeerId int `json:"local_peer_id"`

	// LocalPort is the port the local [comms] module listens to and sends broadcasts on.
	LocalPort int `json:"local_port"`
}

// LoadConfig loads the configuration from a file
func LoadConfig(filename string) *Config {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("[main] Failed to open config file: %v", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := &Config{}
	err = decoder.Decode(config)
	if err != nil {
		log.Fatalf("[main] Failed to decode config file: %v", err)
	}

	return config
}
