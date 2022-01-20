package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	"github.com/leviathan1995/spleen/client/util"
)

type Configuration struct {
	ClientID   int
	ServerIP   string
	ServerPort int
	LimitRate  []string
}

func main() {
	var conf string
	flag.StringVar(&conf, "c", ".spleen-client.json", "The client configuration.")
	flag.Parse()

	bytes, err := ioutil.ReadFile(conf)
	if err != nil {
		log.Fatalf("Reading %s failed.", conf)
	}

	var config Configuration
	if err := json.Unmarshal(bytes, &config); err != nil {
		log.Fatalf("Parsing %s failed.", conf)
	}

	c := client.NewClient(config.ClientID, config.ServerIP, config.ServerPort, config.LimitRate)
	c.Run()
}
