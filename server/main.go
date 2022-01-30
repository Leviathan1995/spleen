package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	"github.com/leviathan1995/spleen/server/util"
)

func main() {
	var conf string
	flag.StringVar(&conf, "c", ".spleen-server.json", "The server configuration")
	flag.Parse()

	bytes, err := ioutil.ReadFile(conf)
	if err != nil {
		log.Fatalf("Reading %s failed.", conf)
	}

	var config server.Configuration
	if err := json.Unmarshal(bytes, &config); err != nil {
		log.Fatalf("Parsing %s failed.", conf)
	}

	s := server.NewServer(config.ServerIP, config.ServerPort, config.Rules)
	s.Listen()
}
