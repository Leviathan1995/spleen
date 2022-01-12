package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	"github.com/leviathan1995/spleen/server/util"
)

type Configuration struct {
	ServerIP    string
	ServerPort  int
	MappingPort []string
}

func main() {
	var conf string
	flag.StringVar(&conf, "c", ".server.json", "The server configuration")
	flag.Parse()

	bytes, err := ioutil.ReadFile(conf)
	if err != nil {
		log.Fatalf("Reading %s failed.", conf)
	}

	var config Configuration
	if err := json.Unmarshal(bytes, &config); err != nil {
		log.Fatalf("Parsing %s failed.", conf)
	}

	s := server.NewServer(config.ServerIP, config.ServerPort, config.MappingPort)
	s.Listen()
}
